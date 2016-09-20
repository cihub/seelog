// Copyright (c) 2013 - Cloud Instruments Co., Ltd.
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package seelog

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Common constants
const (
	rollingLogHistoryDelimiter = "."
)

// Types of the rolling writer: roll by date, by time, etc.
type rollingType uint8

const (
	rollingTypeSize = iota
	rollingTypeTime
)

// Types of the rolled file naming mode: prefix, postfix, etc.
type rollingNameMode uint8

const (
	rollingNameModePostfix = iota
	rollingNameModePrefix
)

var rollingNameModesStringRepresentation = map[rollingNameMode]string{
	rollingNameModePostfix: "postfix",
	rollingNameModePrefix:  "prefix",
}

func rollingNameModeFromString(rollingNameStr string) (rollingNameMode, bool) {
	for tp, tpStr := range rollingNameModesStringRepresentation {
		if tpStr == rollingNameStr {
			return tp, true
		}
	}

	return 0, false
}

var rollingTypesStringRepresentation = map[rollingType]string{
	rollingTypeSize: "size",
	rollingTypeTime: "date",
}

func rollingTypeFromString(rollingTypeStr string) (rollingType, bool) {
	for tp, tpStr := range rollingTypesStringRepresentation {
		if tpStr == rollingTypeStr {
			return tp, true
		}
	}

	return 0, false
}

// Old logs archivation type.
type rollingArchiveType uint8

const (
	rollingArchiveNone = iota
	rollingArchiveZip
	rollingArchiveGzip
)

var rollingArchiveTypesStringRepresentation = map[rollingArchiveType]string{
	rollingArchiveNone: "none",
	rollingArchiveZip:  "zip",
	rollingArchiveGzip: "gzip",
}

type archive func(archiveName string, files map[string]io.Reader, exploded bool) error

type unarchive func(archiveName string) (map[string][]byte, error)

type compressionType struct {
	extension             string
	handleMultipleEntries bool
	archive               archive
	unarchive             unarchive
}

var compressionTypes = map[rollingArchiveType]compressionType{
	rollingArchiveZip: {
		extension:             ".zip",
		handleMultipleEntries: true,
		archive: func(archiveName string, files map[string]io.Reader, exploded bool) error {
			return createZip(archiveName, files)
		},
		unarchive: unzip,
	},
	rollingArchiveGzip: {
		extension:             ".gz",
		handleMultipleEntries: false,
		archive: func(archiveName string, files map[string]io.Reader, exploded bool) error {
			if exploded {
				if len(files) != 1 {
					return fmt.Errorf("Expected only 1 file but got %v file(s)", len(files))
				}
				for _, data := range files {
					return createGzip(archiveName, data)
				}
			}
			tar, err := createTar(files)
			if err != nil {
				return err
			}
			return createGzip(archiveName, bytes.NewReader(tar))
		},
		unarchive: func(archiveName string) (map[string][]byte, error) {
			content, err := unGzip(archiveName)
			if err != nil {
				return nil, err
			}
			if isTar(content) {
				return unTar(content)
			}
			file := make(map[string][]byte)
			file[archiveName] = content
			return file, nil
		},
	},
}

func (compressionType *compressionType) rollingArchiveTypeName(name string, exploded bool) string {
	if !compressionType.handleMultipleEntries && !exploded {
		return name + ".tar" + compressionType.extension
	} else {
		return name + compressionType.extension
	}

}

func rollingArchiveTypeFromString(rollingArchiveTypeStr string) (rollingArchiveType, bool) {
	for tp, tpStr := range rollingArchiveTypesStringRepresentation {
		if tpStr == rollingArchiveTypeStr {
			return tp, true
		}
	}

	return 0, false
}

// Default names for different archive types
var rollingArchiveDefaultExplodedName = "old"

func rollingArchiveTypeDefaultName(archiveType rollingArchiveType, exploded bool) (string, error) {
	compressionType, ok := compressionTypes[archiveType]
	if !ok {
		return "", fmt.Errorf("cannot get default filename for archive type = %v", archiveType)
	}
	return compressionType.rollingArchiveTypeName("log", exploded), nil
}

type rollInfo struct {
	Name string
	Time time.Time
}

// rollerVirtual is an interface that represents all virtual funcs that are
// called in different rolling writer subtypes.
type rollerVirtual interface {
	needsToRoll(lastRollTime time.Time) (bool, error)   // Returns true if needs to switch to another file.
	isFileRollNameValid(rname string) bool              // Returns true if logger roll file name (postfix/prefix/etc.) is ok.
	sortFileRollNamesAsc(fs []string) ([]string, error) // Sorts logger roll file names in ascending order of their creation by logger.

	// Creates a new froll history file using the contents of current file and special filename of the latest roll (prefix/ postfix).
	// If lastRollName is empty (""), then it means that there is no latest roll (current is the first one)
	getNewHistoryRollFileName(lastRoll rollInfo) string

	getCurrentFileName() string
}

// rollingFileWriter writes received messages to a file, until time interval passes
// or file exceeds a specified limit. After that the current log file is renamed
// and writer starts to log into a new file. You can set a limit for such renamed
// files count, if you want, and then the rolling writer would delete older ones when
// the files count exceed the specified limit.
type rollingFileWriter struct {
	fileName        string // log file name
	currentDirPath  string
	currentFile     *os.File
	currentName     string
	currentFileSize int64
	rollingType     rollingType // Rolling mode (Files roll by size/date/...)
	archiveType     rollingArchiveType
	archivePath     string
	archiveExploded bool
	fullName        bool
	maxRolls        int
	nameMode        rollingNameMode
	self            rollerVirtual // Used for virtual calls
}

func newRollingFileWriter(fpath string, rtype rollingType, atype rollingArchiveType, apath string, maxr int, namemode rollingNameMode,
	archiveExploded bool, fullName bool) (*rollingFileWriter, error) {
	rw := new(rollingFileWriter)
	rw.currentDirPath, rw.fileName = filepath.Split(fpath)
	if len(rw.currentDirPath) == 0 {
		rw.currentDirPath = "."
	}

	rw.rollingType = rtype
	rw.archiveType = atype
	rw.archivePath = apath
	rw.nameMode = namemode
	rw.maxRolls = maxr
	rw.archiveExploded = archiveExploded
	rw.fullName = fullName
	return rw, nil
}

func (rw *rollingFileWriter) hasRollName(file string) bool {
	switch rw.nameMode {
	case rollingNameModePostfix:
		rname := rw.fileName + rollingLogHistoryDelimiter
		return strings.HasPrefix(file, rname)
	case rollingNameModePrefix:
		rname := rollingLogHistoryDelimiter + rw.fileName
		return strings.HasSuffix(file, rname)
	}
	return false
}

func (rw *rollingFileWriter) createFullFileName(originalName, rollname string) string {
	switch rw.nameMode {
	case rollingNameModePostfix:
		return originalName + rollingLogHistoryDelimiter + rollname
	case rollingNameModePrefix:
		return rollname + rollingLogHistoryDelimiter + originalName
	}
	return ""
}

func (rw *rollingFileWriter) getSortedLogHistory() ([]string, error) {
	files, err := getDirFilePaths(rw.currentDirPath, nil, true)
	if err != nil {
		return nil, err
	}
	var validRollNames []string
	for _, file := range files {
		if rw.hasRollName(file) {
			rname := rw.getFileRollName(file)
			if rw.self.isFileRollNameValid(rname) {
				validRollNames = append(validRollNames, rname)
			}
		}
	}
	sortedTails, err := rw.self.sortFileRollNamesAsc(validRollNames)
	if err != nil {
		return nil, err
	}
	validSortedFiles := make([]string, len(sortedTails))
	for i, v := range sortedTails {
		validSortedFiles[i] = rw.createFullFileName(rw.fileName, v)
	}
	return validSortedFiles, nil
}

func (rw *rollingFileWriter) createFileAndFolderIfNeeded(first bool) error {
	var err error

	if len(rw.currentDirPath) != 0 {
		err = os.MkdirAll(rw.currentDirPath, defaultDirectoryPermissions)

		if err != nil {
			return err
		}
	}
	rw.currentName = rw.self.getCurrentFileName()
	filePath := filepath.Join(rw.currentDirPath, rw.currentName)

	// If exists
	stat, err := os.Lstat(filePath)
	if err == nil {
		rw.currentFile, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, defaultFilePermissions)
		if err != nil {
			return err
		}

		stat, err = os.Lstat(filePath)
		if err != nil {
			return err
		}

		rw.currentFileSize = stat.Size()
	} else {
		rw.currentFile, err = os.Create(filePath)
		rw.currentFileSize = 0
	}
	if err != nil {
		return err
	}

	return nil
}

func (rw *rollingFileWriter) archiveExplodedLogs(logFilename string, compressionType compressionType) error {
	rollPath := filepath.Join(rw.currentDirPath, logFilename)
	rollFile, err := os.Open(rollPath)
	if err != nil {
		return err
	}
	defer rollFile.Close()

	entry := make(map[string]io.Reader)
	entry[logFilename] = rollFile
	archiveName := path.Clean(rw.archivePath + "/" + compressionType.rollingArchiveTypeName(logFilename, true))

	// archive entry
	return compressionType.archive(archiveName, entry, true)
}

func (rw *rollingFileWriter) archiveUnexplodedLogs(compressionType compressionType, rollsToDelete int, history []string) error {
	var files map[string][]byte
	// If archive exists
	_, err := os.Lstat(rw.archivePath)
	if nil == err {
		// Extract files and content from it
		files, err = compressionType.unarchive(rw.archivePath)
		if err != nil {
			return err
		}

		// Remove the original file
		err = tryRemoveFile(rw.archivePath)
		if err != nil {
			return err
		}
	} else {
		files = make(map[string][]byte)
	}

	// Add files to the existing files map, filled above
	for i := 0; i < rollsToDelete; i++ {
		rollPath := filepath.Join(rw.currentDirPath, history[i])
		bts, err := ioutil.ReadFile(rollPath)
		if err != nil {
			return err
		}

		files[rollPath] = bts
	}

	freaders := make(map[string]io.Reader)
	for fname, fcont := range files {
		freaders[fname] = bytes.NewReader(fcont)
	}

	// Put the final file set to archive file.
	return compressionType.archive(rw.archivePath, freaders, false)
}

func (rw *rollingFileWriter) deleteOldRolls(history []string) error {
	if rw.maxRolls <= 0 {
		return nil
	}

	rollsToDelete := len(history) - rw.maxRolls
	if rollsToDelete <= 0 {
		return nil
	}

	if rw.archiveType != rollingArchiveNone {
		if rw.archiveExploded {
			os.MkdirAll(rw.archivePath, defaultDirectoryPermissions)

			// Archive logs
			for i := 0; i < rollsToDelete; i++ {
				rw.archiveExplodedLogs(history[i], compressionTypes[rw.archiveType])
			}
		} else {
			os.MkdirAll(path.Dir(rw.archivePath), defaultDirectoryPermissions)

			rw.archiveUnexplodedLogs(compressionTypes[rw.archiveType], rollsToDelete, history)
		}
	}

	var err error
	// In all cases (archive files or not) the files should be deleted.
	for i := 0; i < rollsToDelete; i++ {
		// Try best to delete files without breaking the loop.
		if err = tryRemoveFile(filepath.Join(rw.currentDirPath, history[i])); err != nil {
			reportInternalError(err)
		}
	}

	return nil
}

func (rw *rollingFileWriter) getFileRollName(fileName string) string {
	switch rw.nameMode {
	case rollingNameModePostfix:
		return fileName[len(rw.fileName+rollingLogHistoryDelimiter):]
	case rollingNameModePrefix:
		return fileName[:len(fileName)-len(rw.fileName+rollingLogHistoryDelimiter)]
	}
	return ""
}

func (rw *rollingFileWriter) Write(bytes []byte) (n int, err error) {
	if rw.currentFile == nil {
		err := rw.createFileAndFolderIfNeeded(true)
		if err != nil {
			return 0, err
		}
	}
	// needs to roll if:
	//   * file roller max file size exceeded OR
	//   * time roller interval passed
	fi, err := rw.currentFile.Stat()
	if err != nil {
		return 0, err
	}
	lastRollTime := fi.ModTime()
	nr, err := rw.self.needsToRoll(lastRollTime)
	if err != nil {
		return 0, err
	}
	if nr {
		// First, close current file.
		err = rw.currentFile.Close()
		if err != nil {
			return 0, err
		}
		// Current history of all previous log files.
		// For file roller it may be like this:
		//     * ...
		//     * file.log.4
		//     * file.log.5
		//     * file.log.6
		//
		// For date roller it may look like this:
		//     * ...
		//     * file.log.11.Aug.13
		//     * file.log.15.Aug.13
		//     * file.log.16.Aug.13
		// Sorted log history does NOT include current file.
		history, err := rw.getSortedLogHistory()
		if err != nil {
			return 0, err
		}
		// Renames current file to create a new roll history entry
		// For file roller it may be like this:
		//     * ...
		//     * file.log.4
		//     * file.log.5
		//     * file.log.6
		//     n file.log.7  <---- RENAMED (from file.log)
		// Time rollers that doesn't modify file names (e.g. 'date' roller) skip this logic.
		var newHistoryName string
		lastRoll := rollInfo{
			Time: lastRollTime,
		}
		if len(history) > 0 {
			// Create new rname name using last history file name
			lastRoll.Name = rw.getFileRollName(history[len(history)-1])
		} else {
			// Create first rname name
			lastRoll.Name = ""
		}
		newRollMarkerName := rw.self.getNewHistoryRollFileName(lastRoll)
		if len(newRollMarkerName) != 0 {
			newHistoryName = rw.createFullFileName(rw.fileName, newRollMarkerName)
		} else {
			newHistoryName = rw.fileName
		}
		if newHistoryName != rw.fileName {
			err = os.Rename(filepath.Join(rw.currentDirPath, rw.currentName), filepath.Join(rw.currentDirPath, newHistoryName))
			if err != nil {
				return 0, err
			}
		}
		// Finally, add the newly added history file to the history archive
		// and, if after that the archive exceeds the allowed max limit, older rolls
		// must the removed/archived.
		history = append(history, newHistoryName)
		if len(history) > rw.maxRolls {
			err = rw.deleteOldRolls(history)
			if err != nil {
				return 0, err
			}
		}

		err = rw.createFileAndFolderIfNeeded(false)
		if err != nil {
			return 0, err
		}
	}

	rw.currentFileSize += int64(len(bytes))
	return rw.currentFile.Write(bytes)
}

func (rw *rollingFileWriter) Close() error {
	if rw.currentFile != nil {
		e := rw.currentFile.Close()
		if e != nil {
			return e
		}
		rw.currentFile = nil
	}
	return nil
}

// =============================================================================================
//      Different types of rolling writers
// =============================================================================================

// --------------------------------------------------
//      Rolling writer by SIZE
// --------------------------------------------------

// rollingFileWriterSize performs roll when file exceeds a specified limit.
type rollingFileWriterSize struct {
	*rollingFileWriter
	maxFileSize int64
}

func NewRollingFileWriterSize(fpath string, atype rollingArchiveType, apath string, maxSize int64, maxRolls int, namemode rollingNameMode, archiveExploded bool) (*rollingFileWriterSize, error) {
	rw, err := newRollingFileWriter(fpath, rollingTypeSize, atype, apath, maxRolls, namemode, archiveExploded, false)
	if err != nil {
		return nil, err
	}
	rws := &rollingFileWriterSize{rw, maxSize}
	rws.self = rws
	return rws, nil
}

func (rws *rollingFileWriterSize) needsToRoll(lastRollTime time.Time) (bool, error) {
	return rws.currentFileSize >= rws.maxFileSize, nil
}

func (rws *rollingFileWriterSize) isFileRollNameValid(rname string) bool {
	if len(rname) == 0 {
		return false
	}
	_, err := strconv.Atoi(rname)
	return err == nil
}

type rollSizeFileTailsSlice []string

func (p rollSizeFileTailsSlice) Len() int {
	return len(p)
}
func (p rollSizeFileTailsSlice) Less(i, j int) bool {
	v1, _ := strconv.Atoi(p[i])
	v2, _ := strconv.Atoi(p[j])
	return v1 < v2
}
func (p rollSizeFileTailsSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (rws *rollingFileWriterSize) sortFileRollNamesAsc(fs []string) ([]string, error) {
	ss := rollSizeFileTailsSlice(fs)
	sort.Sort(ss)
	return ss, nil
}

func (rws *rollingFileWriterSize) getNewHistoryRollFileName(lastRoll rollInfo) string {
	v := 0
	if len(lastRoll.Name) != 0 {
		v, _ = strconv.Atoi(lastRoll.Name)
	}
	return fmt.Sprintf("%d", v+1)
}

func (rws *rollingFileWriterSize) getCurrentFileName() string {
	return rws.fileName
}

func (rws *rollingFileWriterSize) String() string {
	return fmt.Sprintf("Rolling file writer (By SIZE): filename: %s, archive: %s, archivefile: %s, maxFileSize: %v, maxRolls: %v",
		rws.fileName,
		rollingArchiveTypesStringRepresentation[rws.archiveType],
		rws.archivePath,
		rws.maxFileSize,
		rws.maxRolls)
}

// --------------------------------------------------
//      Rolling writer by TIME
// --------------------------------------------------

// rollingFileWriterTime performs roll when a specified time interval has passed.
type rollingFileWriterTime struct {
	*rollingFileWriter
	timePattern         string
	currentTimeFileName string
}

func NewRollingFileWriterTime(fpath string, atype rollingArchiveType, apath string, maxr int,
	timePattern string, namemode rollingNameMode, archiveExploded bool, fullName bool) (*rollingFileWriterTime, error) {

	rw, err := newRollingFileWriter(fpath, rollingTypeTime, atype, apath, maxr, namemode, archiveExploded, fullName)
	if err != nil {
		return nil, err
	}
	rws := &rollingFileWriterTime{rw, timePattern, ""}
	rws.self = rws
	return rws, nil
}

func (rwt *rollingFileWriterTime) needsToRoll(lastRollTime time.Time) (bool, error) {
	if time.Now().Format(rwt.timePattern) == lastRollTime.Format(rwt.timePattern) {
		return false, nil
	}
	return true, nil
}

func (rwt *rollingFileWriterTime) isFileRollNameValid(rname string) bool {
	if len(rname) == 0 {
		return false
	}
	_, err := time.ParseInLocation(rwt.timePattern, rname, time.Local)
	return err == nil
}

type rollTimeFileTailsSlice struct {
	data    []string
	pattern string
}

func (p rollTimeFileTailsSlice) Len() int {
	return len(p.data)
}

func (p rollTimeFileTailsSlice) Less(i, j int) bool {
	t1, _ := time.ParseInLocation(p.pattern, p.data[i], time.Local)
	t2, _ := time.ParseInLocation(p.pattern, p.data[j], time.Local)
	return t1.Before(t2)
}

func (p rollTimeFileTailsSlice) Swap(i, j int) {
	p.data[i], p.data[j] = p.data[j], p.data[i]
}

func (rwt *rollingFileWriterTime) sortFileRollNamesAsc(fs []string) ([]string, error) {
	ss := rollTimeFileTailsSlice{data: fs, pattern: rwt.timePattern}
	sort.Sort(ss)
	return ss.data, nil
}

func (rwt *rollingFileWriterTime) getNewHistoryRollFileName(lastRoll rollInfo) string {
	return lastRoll.Time.Format(rwt.timePattern)
}

func (rwt *rollingFileWriterTime) getCurrentFileName() string {
	if rwt.fullName {
		return rwt.createFullFileName(rwt.fileName, time.Now().Format(rwt.timePattern))
	}
	return rwt.fileName
}

func (rwt *rollingFileWriterTime) String() string {
	return fmt.Sprintf("Rolling file writer (By TIME): filename: %s, archive: %s, archivefile: %s, pattern: %s, maxRolls: %v",
		rwt.fileName,
		rollingArchiveTypesStringRepresentation[rwt.archiveType],
		rwt.archivePath,
		rwt.timePattern,
		rwt.maxRolls)
}
