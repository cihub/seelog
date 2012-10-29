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
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type rollingTypes uint8

const (
	Size = iota
	Date
)

var rollingTypesStringRepresentation = map[rollingTypes]string{
	Size: "size",
	Date: "date",
}

func rollingTypeFromString(rollingTypeStr string) (rollingType rollingTypes, found bool) {
	for tp, tpStr := range rollingTypesStringRepresentation {
		if tpStr == rollingTypeStr {
			return tp, true
		}
	}

	return 0, false
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
	currentFilePath string
	currentFileSize int64
	innerWriter     io.WriteCloser // Represents file
}

// newRollingFileWriterSize initializes a rolling writer with a 'Size' rolling mode
func newRollingFileWriterSize(filePath string, maxFileSize int64, maxRolls int) (*rollingFileWriter, error) {
	if maxFileSize <= 0 {
		return nil, errors.New("maxFileSize must be positive")
	}

	if maxRolls <= 0 {
		return nil, errors.New("maxFileSize must be positive")
	}

	rollingFile := new(rollingFileWriter)
	rollingFile.rollingType = Size
	rollingFile.maxFileSize = maxFileSize
	rollingFile.maxRolls = maxRolls
	rollingFile.filePath = filePath
	rollingFile.fileDir, rollingFile.fileName = filepath.Split(filePath)

	return rollingFile, nil
}

// newRollingFileWriterSize initializes a rolling writer with a 'Date' rolling mode
func newRollingFileWriterDate(filePath string, datePattern string) (*rollingFileWriter, error) {
	rollingFile := new(rollingFileWriter)
	rollingFile.rollingType = Date
	rollingFile.datePattern = datePattern
	rollingFile.filePath = filePath
	rollingFile.fileDir, rollingFile.fileName = filepath.Split(filePath)

	return rollingFile, nil
}

func (rollfileWriter *rollingFileWriter) getFileName() string {
	if rollfileWriter.rollingType == Size {
		return rollfileWriter.fileName
	} else if rollfileWriter.rollingType == Date {
		return time.Now().Format(rollfileWriter.datePattern) + " " + rollfileWriter.fileName
	}

	return rollfileWriter.fileName
}

func (rollfileWriter *rollingFileWriter) isTimeToCreateFile() bool {
	if rollfileWriter.innerWriter == nil {
		return true
	}

	if rollfileWriter.rollingType == Size {
		return rollfileWriter.currentFileSize >= rollfileWriter.maxFileSize
	} else if rollfileWriter.rollingType == Date {
		fileName := rollfileWriter.getFileName()
		return rollfileWriter.currentFileName != fileName
	}

	return false
}

func (rollfileWriter *rollingFileWriter) createFile() error {
	if rollfileWriter.innerWriter == nil {
		return rollfileWriter.createFileAndFolderIfNeeded()
	}

	if rollfileWriter.rollingType == Size {
		if rollfileWriter.innerWriter != nil {
			rollfileWriter.innerWriter.Close()
		}

		nextRollName, err := rollfileWriter.getNextRollName()
		if err != nil {
			return err
		}

		err = fileSystemWrapper.Rename(rollfileWriter.currentFilePath, filepath.Join(rollfileWriter.fileDir, nextRollName))
		if err != nil {
			return err
		}

		rollfileWriter.deleteOldRolls()

		return rollfileWriter.createFileAndFolderIfNeeded()
	} else if rollfileWriter.rollingType == Date {
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
	files, err := fileSystemWrapper.GetDirFileNames(rollfileWriter.fileDir, false)
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
	for i := 0; i < rollsToDelete; i++ {
		err := fileSystemWrapper.Remove(filepath.Join(rollfileWriter.fileDir, sortedRolls[i]))
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

func (rollfileWriter *rollingFileWriter) Close() error {
	return rollfileWriter.innerWriter.Close()
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

func (rollfileWriter *rollingFileWriter) createFileAndFolderIfNeeded() error {
	err := fileSystemWrapper.MkdirAll(rollfileWriter.fileDir)
	if err != nil {
		return err
	}

	if rollfileWriter.innerWriter != nil {
		rollfileWriter.innerWriter.Close()
	}

	fileName := rollfileWriter.getFileName()
	filePath := filepath.Join(rollfileWriter.fileDir, fileName)

	var innerWriter io.WriteCloser
	if fileSystemWrapper.Exists(filePath) {
		innerWriter, err = fileSystemWrapper.Open(filePath)
		size, err := fileSystemWrapper.GetFileSize(filePath)
		if err != nil {
			return err
		}
		rollfileWriter.currentFileSize = size
	} else {
		innerWriter, err = fileSystemWrapper.Create(filePath)
		rollfileWriter.currentFileSize = 0
	}
	if err != nil {
		return err
	}

	rollfileWriter.currentFilePath = filePath
	rollfileWriter.currentFileName = fileName
	rollfileWriter.innerWriter = innerWriter

	return nil
}

func (rollfileWriter *rollingFileWriter) String() string {

	rollingTypeStr, ok := rollingTypesStringRepresentation[rollfileWriter.rollingType]
	if !ok {
		rollingTypeStr = "UNKNOWN"
	}

	s := fmt.Sprintf("Rolling file writer: filename: %s type: %s ", rollfileWriter.fileName, rollingTypeStr)

	if rollfileWriter.rollingType == Size {
		s += fmt.Sprintf("maxFileSize: %v, maxRolls: %v", rollfileWriter.maxFileSize, rollfileWriter.maxRolls)
	} else if rollfileWriter.rollingType == Date {
		s += fmt.Sprintf("datePattern: %v", rollfileWriter.datePattern)
	}

	return s
}
