// Copyright (c) 2012 - Cloud Instruments Co., Ltd.
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
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Types of the rolling writer: roll by date, by time, etc.
type rollingTypes uint8

const (
	rollingTypeSize = iota
	rollingTypeDate
)

var rollingTypesStringRepresentation = map[rollingTypes]string{
	rollingTypeSize: "size",
	rollingTypeDate: "date",
}

func rollingTypeFromString(rollingTypeStr string) (rollingType rollingTypes, found bool) {
	for tp, tpStr := range rollingTypesStringRepresentation {
		if tpStr == rollingTypeStr {
			return tp, true
		}
	}

	return 0, false
}

// Old logs archivation type.
type rollingArchiveTypes uint8

const (
	rollingArchiveNone = iota
	rollingArchiveZip
)

var rollingArchiveTypesStringRepresentation = map[rollingArchiveTypes]string{
	rollingArchiveNone: "none",
	rollingArchiveZip:  "zip",
}

func rollingArchiveTypeFromString(rollingArchiveTypeStr string) (rollingArchiveType rollingArchiveTypes, found bool) {
	for tp, tpStr := range rollingArchiveTypesStringRepresentation {
		if tpStr == rollingArchiveTypeStr {
			return tp, true
		}
	}

	return 0, false
}

// Default names for different archivation types
var rollingArchiveTypesDefaultNames = map[rollingArchiveTypes]string{
	rollingArchiveZip: "log.zip",
}

// rollingFileWriter writes received messages to a file, until date changes
// or file exceeds a specified limit. After that the current log file is renamed
// and writer starts to log into a new file. You can set a limit for such renamed
// files count, if you want, and then the rolling writer would delete older ones when
// the files count exceed the specified limit
type rollingFileWriter struct {
	fileName    string
	fileDir     string // Rolling files folder
	filePath    string
	rollingType rollingTypes // Rolling mode (Files roll by size/date/...)

	maxFileSize int64 // Maximal file size at which roll must occur
	maxRolls    int   // Maximal count of roll files that exist at the same time

	datePattern string // DateTime pattern used as roll files prefix

	currentFileName string
	currentFileSize int64
	innerWriter     io.WriteCloser // Represents file

	archiveType rollingArchiveTypes
	archivePath string
}

// newRollingFileWriterSize initializes a rolling writer with a 'Size' rolling mode
func newRollingFileWriterSize(
	filePath string,
	arch rollingArchiveTypes,
	archPath string,
	maxFileSize int64,
	maxRolls int) (*rollingFileWriter, error) {

	if maxFileSize <= 0 {
		return nil, errors.New("maxFileSize must be positive")
	}

	if maxRolls <= 0 {
		return nil, errors.New("maxFileSize must be positive")
	}

	rollingFile := new(rollingFileWriter)

	rollingFile.archiveType = arch
	rollingFile.rollingType = rollingTypeSize
	rollingFile.maxFileSize = maxFileSize
	rollingFile.maxRolls = maxRolls
	rollingFile.filePath = filePath
	rollingFile.fileDir, rollingFile.fileName = filepath.Split(filePath)
	rollingFile.archivePath = archPath

	return rollingFile, nil
}

// newRollingFileWriterSize initializes a rolling writer with a 'Date' rolling mode
func newRollingFileWriterDate(
	filePath string,
	arch rollingArchiveTypes,
	archPath string,
	datePattern string) (*rollingFileWriter, error) {

	rollingFile := new(rollingFileWriter)

	rollingFile.archiveType = arch
	rollingFile.rollingType = rollingTypeDate
	rollingFile.datePattern = datePattern
	rollingFile.filePath = filePath
	rollingFile.fileDir, rollingFile.fileName = filepath.Split(filePath)
	rollingFile.archivePath = archPath

	return rollingFile, nil
}

func (rollfileWriter *rollingFileWriter) getFileName() string {
	if rollfileWriter.rollingType == rollingTypeSize {
		return rollfileWriter.fileName
	} else if rollfileWriter.rollingType == rollingTypeDate {
		return time.Now().Format(rollfileWriter.datePattern) + " " + rollfileWriter.fileName
	}

	return rollfileWriter.fileName
}

func (rollfileWriter *rollingFileWriter) isTimeToCreateFile() bool {
	if rollfileWriter.innerWriter == nil {
		return true
	}

	if rollfileWriter.rollingType == rollingTypeSize {
		return rollfileWriter.currentFileSize >= rollfileWriter.maxFileSize
	} else if rollfileWriter.rollingType == rollingTypeDate {
		fileName := rollfileWriter.getFileName()
		return rollfileWriter.currentFileName != fileName
	}

	return false
}

func (rollfileWriter *rollingFileWriter) createFile() error {
	if rollfileWriter.innerWriter == nil {
		return rollfileWriter.createFileAndFolderIfNeeded()
	}

	e := rollfileWriter.Close()

	if e != nil {
		return e
	}

	if rollfileWriter.rollingType == rollingTypeSize {

		nextRollName, err := rollfileWriter.getNextRollName()
		if err != nil {
			return err
		}

		currentFilePath := filepath.Join(rollfileWriter.fileDir, rollfileWriter.currentFileName)
		nextFilePath := filepath.Join(rollfileWriter.fileDir, nextRollName)

		err = os.Rename(currentFilePath, nextFilePath)
		if err != nil {
			return err
		}

		rollfileWriter.deleteOldRolls()

		return rollfileWriter.createFileAndFolderIfNeeded()
	} else if rollfileWriter.rollingType == rollingTypeDate {
		return rollfileWriter.createFileAndFolderIfNeeded()
	}

	return nil
}

func (rollfileWriter *rollingFileWriter) getNextRollName() (string, error) {
	rolls, err := rollfileWriter.getRolls()
	if err != nil {
		return "", err
	}

	var nextIndex = 1
	for index, _ := range rolls {
		if index >= nextIndex {
			nextIndex = index + 1
		}
	}

	return rollfileWriter.currentFileName + "." + strconv.Itoa(nextIndex), nil
}

func (rollfileWriter *rollingFileWriter) getRolls() (map[int]string, error) {
	var dir string

	if len(rollfileWriter.fileDir) == 0 {
		dir = "."
	} else {
		dir = rollfileWriter.fileDir
	}

	files, err := getDirFilePaths(dir, nil, true)

	if err != nil {
		return map[int]string{}, err
	}

	rolls := make(map[int]string, 0)

	for _, file := range files {
		if strings.HasPrefix(file, rollfileWriter.currentFileName) {
			if len(rollfileWriter.currentFileName)+1 >= len(file) {
				continue
			}

			fileIndex := file[len(rollfileWriter.currentFileName)+1:]
			index, err := strconv.Atoi(fileIndex)
			if err != nil {
				continue
			}

			rolls[index] = file
		}
	}

	return rolls, nil
}

// Unzips a specified zip file. Returns filename->filebytes map.
func unzip(archiveName string) (map[string][]byte, error) {
	// Open a zip archive for reading.
	r, err := zip.OpenReader(archiveName)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Files to be added to archive
	// map file name to contents
	files := make(map[string][]byte)

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}

		bts, err := ioutil.ReadAll(rc)
		rcErr := rc.Close()

		if err != nil {
			return nil, err
		}
		if rcErr != nil {
			return nil, rcErr
		}

		files[f.Name] = bts
	}

	return files, nil
}

// Creates a zip file with the specified file names and byte contents.
func createZip(archiveName string, files map[string][]byte) error {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	// Write files
	for fpath, fcont := range files {
		f, err := w.Create(fpath)
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(fcont))
		if err != nil {
			return err
		}
	}

	// Make sure to check the error on Close.
	err := w.Close()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(archiveName, buf.Bytes(), defaultFilePermissions)
	if err != nil {
		return err
	}

	return nil
}

func (rollfileWriter *rollingFileWriter) deleteOldRolls() error {
	if rollfileWriter.maxRolls <= 0 {
		return nil
	}

	rolls, err := rollfileWriter.getRolls()
	if err != nil {
		return err
	}

	rollsToDelete := len(rolls) - rollfileWriter.maxRolls
	if rollsToDelete <= 0 {
		return nil
	}

	sortedRolls := rollfileWriter.sortRollsByIndex(rolls)

	switch rollfileWriter.archiveType {
	case rollingArchiveZip:
		var files map[string][]byte

		// If archive exists
		_, err := os.Lstat(rollfileWriter.archivePath)
		if nil == err {
			// Extract files and content from it
			files, err = unzip(rollfileWriter.archivePath)
			if err != nil {
				return err
			}

			// Remove the original file
			err = tryRemoveFile(rollfileWriter.archivePath)
			if err != nil {
				return err
			}
		} else {
			files = make(map[string][]byte)
		}

		// Add files to the existing files map, filled above
		for i := 0; i < rollsToDelete; i++ {
			rollPath := filepath.Join(rollfileWriter.fileDir, sortedRolls[i])
			bts, err := ioutil.ReadFile(rollPath)
			if err != nil {
				return err
			}

			files[rollPath] = bts
		}

		// Put the final file set to zip file.
		err = createZip(rollfileWriter.archivePath, files)
		if err != nil {
			return err
		}
	}

	// In all cases (archive files or not) the files should be deleted.
	for i := 0; i < rollsToDelete; i++ {
		rollPath := filepath.Join(rollfileWriter.fileDir, sortedRolls[i])
		err := tryRemoveFile(rollPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rollfileWriter *rollingFileWriter) sortRollsByIndex(rolls map[int]string) []string {
	indexes := make([]int, 0)
	for index, _ := range rolls {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)

	sortedRolls := make([]string, len(indexes))
	for i, index := range indexes {
		sortedRolls[i] = rolls[index]
	}
	return sortedRolls
}

func (rollfileWriter *rollingFileWriter) createFileAndFolderIfNeeded() error {
	var err error

	if 0 != len(rollfileWriter.fileDir) {
		err = os.MkdirAll(rollfileWriter.fileDir, defaultDirectoryPermissions)

		if err != nil {
			return err
		}
	}

	if rollfileWriter.innerWriter != nil {
		err = rollfileWriter.innerWriter.Close()

		if err != nil {
			return err
		}
	}

	fileName := rollfileWriter.getFileName()
	filePath := filepath.Join(rollfileWriter.fileDir, fileName)

	// If exists
	stat, err := os.Lstat(filePath)
	if err == nil {
		rollfileWriter.innerWriter, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, defaultFilePermissions)

		stat, err = os.Lstat(filePath)
		if err != nil {
			return err
		}

		rollfileWriter.currentFileSize = stat.Size()
	} else {
		rollfileWriter.innerWriter, err = os.Create(filePath)
		rollfileWriter.currentFileSize = 0
	}
	if err != nil {
		return err
	}

	rollfileWriter.currentFileName = fileName

	return nil
}

func (rollfileWriter *rollingFileWriter) String() string {

	rollingTypeStr, ok := rollingTypesStringRepresentation[rollfileWriter.rollingType]
	if !ok {
		rollingTypeStr = "UNKNOWN"
	}

	rollingArchiveTypeStr, ok := rollingArchiveTypesStringRepresentation[rollfileWriter.archiveType]
	if !ok {
		rollingArchiveTypeStr = "UNKNOWN"
	}

	s := fmt.Sprintf("Rolling file writer: filename: %s type: %s archive: %s archivefile: %s",
		rollfileWriter.fileName,
		rollingTypeStr,
		rollingArchiveTypeStr,
		rollfileWriter.archivePath)

	if rollfileWriter.rollingType == rollingTypeSize {
		s += fmt.Sprintf("maxFileSize: %v, maxRolls: %v", rollfileWriter.maxFileSize, rollfileWriter.maxRolls)
	} else if rollfileWriter.rollingType == rollingTypeDate {
		s += fmt.Sprintf("datePattern: %v", rollfileWriter.datePattern)
	}

	return s
}

func (rollfileWriter *rollingFileWriter) Close() error {
	if rollfileWriter.innerWriter != nil {
		e := rollfileWriter.innerWriter.Close()

		if e != nil {
			return e
		}

		rollfileWriter.innerWriter = nil
	}
	return nil
}

func (rollfileWriter *rollingFileWriter) Write(bytes []byte) (n int, err error) {
	if rollfileWriter.isTimeToCreateFile() {
		err := rollfileWriter.createFile()
		if err != nil {
			return 0, err
		}
	}

	if rollfileWriter.innerWriter != nil {
		rollfileWriter.currentFileSize += int64(len(bytes))
		return rollfileWriter.innerWriter.Write(bytes)
	}

	return 0, nil
}
