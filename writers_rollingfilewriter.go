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
	"fmt"
	"io/ioutil"
	"os"
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

type rollingIntervalType uint8

const (
	rollingIntervalAny = iota
	rollingIntervalDaily
)

var rollingInvervalTypesStringRepresentation = map[rollingIntervalType]string{
	rollingIntervalDaily: "daily",
}

func rollingIntervalTypeFromString(rollingTypeStr string) (rollingIntervalType, bool) {
	for tp, tpStr := range rollingInvervalTypesStringRepresentation {
		if tpStr == rollingTypeStr {
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
)

var rollingArchiveTypesStringRepresentation = map[rollingArchiveType]string{
	rollingArchiveNone: "none",
	rollingArchiveZip:  "zip",
}

func rollingArchiveTypeFromString(rollingArchiveTypeStr string) (rollingArchiveType, bool) {
	for tp, tpStr := range rollingArchiveTypesStringRepresentation {
		if tpStr == rollingArchiveTypeStr {
			return tp, true
		}
	}

	return 0, false
}

// Default names for different archivation types
var rollingArchiveTypesDefaultNames = map[rollingArchiveType]string{
	rollingArchiveZip: "log.zip",
}

// rollerVirtual is an interface that represents all virtual funcs that are
// called in different rolling writer subtypes.
type rollerVirtual interface {
	needsToRoll() (bool, error)                     // Returns true if needs to switch to another file.
	isFileTailValid(tail string) bool               // Returns true if logger roll file tail (part after filename) is ok.
	sortFileTailsAsc(fs []string) ([]string, error) // Sorts logger roll file tails in ascending order of their creation by logger.

	// Creates a new froll history file using the contents of current file and filename of the latest roll.
	// If lastRollFileTail is empty (""), then it means that there is no latest roll (current is the first one)
	getNewHistoryFileNameTail(lastRollFileTail string) string
	getCurrentModifiedFileName(originalFileName string) string // Returns filename modified according to specific logger rules
}

// rollingFileWriter writes received messages to a file, until time interval passes
// or file exceeds a specified limit. After that the current log file is renamed
// and writer starts to log into a new file. You can set a limit for such renamed
// files count, if you want, and then the rolling writer would delete older ones when
// the files count exceed the specified limit.
type rollingFileWriter struct {
	fileName         string // current file name. May differ from original in date rolling loggers
	originalFileName string // original one
	currentDirPath   string
	currentFile      *os.File
	currentFileSize  int64
	rollingType      rollingType // Rolling mode (Files roll by size/date/...)
	archiveType      rollingArchiveType
	archivePath      string
	maxRolls         int
	self             rollerVirtual // Used for virtual calls
}

func newRollingFileWriter(fpath string, rtype rollingType, atype rollingArchiveType, apath string, maxr int) (*rollingFileWriter, error) {
	rw := new(rollingFileWriter)
	rw.currentDirPath, rw.fileName = filepath.Split(fpath)
	if len(rw.currentDirPath) == 0 {
		rw.currentDirPath = "."
	}
	rw.originalFileName = rw.fileName

	rw.rollingType = rtype
	rw.archiveType = atype
	rw.archivePath = apath
	rw.maxRolls = maxr
	return rw, nil
}

func (rw *rollingFileWriter) getSortedLogHistory() ([]string, error) {
	files, err := getDirFilePaths(rw.currentDirPath, nil, true)
	if err != nil {
		return nil, err
	}
	pref := rw.originalFileName + rollingLogHistoryDelimiter
	var validFileTails []string
	for _, file := range files {
		if file != rw.fileName && strings.HasPrefix(file, pref) {
			tail := rw.getFileTail(file)
			if rw.self.isFileTailValid(tail) {
				validFileTails = append(validFileTails, tail)
			}
		}
	}
	sortedTails, err := rw.self.sortFileTailsAsc(validFileTails)
	if err != nil {
		return nil, err
	}
	validSortedFiles := make([]string, len(sortedTails))
	for i, v := range sortedTails {
		validSortedFiles[i] = rw.originalFileName + rollingLogHistoryDelimiter + v
	}
	return validSortedFiles, nil
}

func (rw *rollingFileWriter) createFileAndFolderIfNeeded() error {
	var err error

	if len(rw.currentDirPath) != 0 {
		err = os.MkdirAll(rw.currentDirPath, defaultDirectoryPermissions)

		if err != nil {
			return err
		}
	}

	rw.fileName = rw.self.getCurrentModifiedFileName(rw.originalFileName)
	filePath := filepath.Join(rw.currentDirPath, rw.fileName)

	// If exists
	stat, err := os.Lstat(filePath)
	if err == nil {
		rw.currentFile, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, defaultFilePermissions)

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

func (rw *rollingFileWriter) deleteOldRolls(history []string) error {
	if rw.maxRolls <= 0 {
		return nil
	}

	rollsToDelete := len(history) - rw.maxRolls
	if rollsToDelete <= 0 {
		return nil
	}

	switch rw.archiveType {
	case rollingArchiveZip:
		var files map[string][]byte

		// If archive exists
		_, err := os.Lstat(rw.archivePath)
		if nil == err {
			// Extract files and content from it
			files, err = unzip(rw.archivePath)
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

		// Put the final file set to zip file.
		err = createZip(rw.archivePath, files)
		if err != nil {
			return err
		}
	}

	// In all cases (archive files or not) the files should be deleted.
	for i := 0; i < rollsToDelete; i++ {
		rollPath := filepath.Join(rw.currentDirPath, history[i])
		err := tryRemoveFile(rollPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rw *rollingFileWriter) getFileTail(fileName string) string {
	return fileName[len(rw.originalFileName+rollingLogHistoryDelimiter):]
}

func (rw *rollingFileWriter) Write(bytes []byte) (n int, err error) {
	if rw.currentFile == nil {
		err := rw.createFileAndFolderIfNeeded()
		if err != nil {
			return 0, err
		}
	}
	// needs to roll if:
	//   * file roller max file size exceeded OR
	//   * time roller interval passed
	nr, err := rw.self.needsToRoll()
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
		var newTail string
		if len(history) > 0 {
			// Create new tail name using last history file name
			newTail = rw.self.getNewHistoryFileNameTail(rw.getFileTail(history[len(history)-1]))
		} else {
			// Create first tail name
			newTail = rw.self.getNewHistoryFileNameTail("")
		}

		if len(newTail) != 0 {
			newHistoryName = rw.fileName + rollingLogHistoryDelimiter + newTail
		} else {
			newHistoryName = rw.fileName
		}

		if newHistoryName != rw.fileName {
			err = os.Rename(filepath.Join(rw.currentDirPath, rw.fileName), filepath.Join(rw.currentDirPath, newHistoryName))
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

		err = rw.createFileAndFolderIfNeeded()
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

func newRollingFileWriterSize(fpath string, atype rollingArchiveType, apath string, maxSize int64, maxRolls int) (*rollingFileWriterSize, error) {
	rw, err := newRollingFileWriter(fpath, rollingTypeSize, atype, apath, maxRolls)
	if err != nil {
		return nil, err
	}
	rws := &rollingFileWriterSize{rw, maxSize}
	rws.self = rws
	return rws, nil
}

func (rws *rollingFileWriterSize) needsToRoll() (bool, error) {
	return rws.currentFileSize >= rws.maxFileSize, nil
}

func (rws *rollingFileWriterSize) isFileTailValid(tail string) bool {
	if len(tail) == 0 {
		return false
	}
	_, err := strconv.Atoi(tail)
	return err == nil
}

type rollSizeFileTailsSlice []string

func (p rollSizeFileTailsSlice) Len() int { return len(p) }
func (p rollSizeFileTailsSlice) Less(i, j int) bool {
	v1, _ := strconv.Atoi(p[i])
	v2, _ := strconv.Atoi(p[j])
	return v1 < v2
}
func (p rollSizeFileTailsSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (rws *rollingFileWriterSize) sortFileTailsAsc(fs []string) ([]string, error) {
	ss := rollSizeFileTailsSlice(fs)
	sort.Sort(ss)
	return ss, nil
}

func (rws *rollingFileWriterSize) getNewHistoryFileNameTail(lastRollFileTail string) string {
	v := 0
	if len(lastRollFileTail) != 0 {
		v, _ = strconv.Atoi(lastRollFileTail)
	}
	return fmt.Sprintf("%d", v+1)
}

func (rws *rollingFileWriterSize) getCurrentModifiedFileName(originalFileName string) string {
	return originalFileName
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
	interval            rollingIntervalType
	currentTimeFileName string
}

func newRollingFileWriterTime(fpath string, atype rollingArchiveType, apath string, maxr int,
	timePattern string, interval rollingIntervalType) (*rollingFileWriterTime, error) {

	rw, err := newRollingFileWriter(fpath, rollingTypeTime, atype, apath, maxr)
	if err != nil {
		return nil, err
	}
	rws := &rollingFileWriterTime{rw, timePattern, interval, ""}
	rws.self = rws
	return rws, nil
}

func (rwt *rollingFileWriterTime) needsToRoll() (bool, error) {
	if rwt.originalFileName+rollingLogHistoryDelimiter+time.Now().Format(rwt.timePattern) == rwt.fileName {
		return false, nil
	}
	if rwt.interval == rollingIntervalAny {
		return true, nil
	}

	tprev, err := time.ParseInLocation(rwt.timePattern, rwt.getFileTail(rwt.fileName), time.Local)
	if err != nil {
		return false, err
	}

	diff := time.Now().Sub(tprev)
	switch rwt.interval {
	case rollingIntervalDaily:
		return diff >= 24*time.Hour, nil
	}
	return false, fmt.Errorf("Unknown interval type: %d", rwt.interval)
}

func (rwt *rollingFileWriterTime) isFileTailValid(tail string) bool {
	if len(tail) == 0 {
		return false
	}
	_, err := time.ParseInLocation(rwt.timePattern, tail, time.Local)
	return err == nil
}

type rollTimeFileTailsSlice struct {
	data    []string
	pattern string
}

func (p rollTimeFileTailsSlice) Len() int { return len(p.data) }
func (p rollTimeFileTailsSlice) Less(i, j int) bool {
	t1, _ := time.ParseInLocation(p.pattern, p.data[i], time.Local)
	t2, _ := time.ParseInLocation(p.pattern, p.data[j], time.Local)
	return t1.Before(t2)
}
func (p rollTimeFileTailsSlice) Swap(i, j int) { p.data[i], p.data[j] = p.data[j], p.data[i] }

func (rwt *rollingFileWriterTime) sortFileTailsAsc(fs []string) ([]string, error) {
	ss := rollTimeFileTailsSlice{data: fs, pattern: rwt.timePattern}
	sort.Sort(ss)
	return ss.data, nil
}

func (rwt *rollingFileWriterTime) getNewHistoryFileNameTail(lastRollFileTail string) string {
	return ""
}

func (rwt *rollingFileWriterTime) getCurrentModifiedFileName(originalFileName string) string {
	return originalFileName + rollingLogHistoryDelimiter + time.Now().Format(rwt.timePattern)
}

func (rwt *rollingFileWriterTime) String() string {
	return fmt.Sprintf("Rolling file writer (By TIME): filename: %s, archive: %s, archivefile: %s, maxInterval: %v, pattern: %s, maxRolls: %v",
		rwt.fileName,
		rollingArchiveTypesStringRepresentation[rwt.archiveType],
		rwt.archivePath,
		rwt.interval,
		rwt.timePattern,
		rwt.maxRolls)
}
