// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package writers

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

type RollingTypes uint8

const (
	Size = iota
	Date
)

var rollingTypesStringRepresentation = map[RollingTypes]string{
	Size: "size",
	Date: "date",
}

func RollingTypeFromString(rollingTypeStr string) (rollingType RollingTypes, found bool) {
	for tp, tpStr := range rollingTypesStringRepresentation {
		if tpStr == rollingTypeStr {
			return tp, true
		}
	}

	return 0, false
}

// RollingFileWriter writes received messages to a file, until date changes
// or file exceeds a specified limit. After that the current log file is renamed 
// and writer starts to log into a new file. You can set a limit for such renamed 
// files count, if you want, and then the rolling writer would delete older ones when
// the files count exceed the specified limit
type RollingFileWriter struct {
	fileName    string
	fileDir     string       // Rolling files folder
	filePath    string
	rollingType RollingTypes // Rolling mode (Files roll by size/date/...)

	maxFileSize int64 // Maximal file size at which roll must occur
	maxRolls    int   // Maximal count of roll files that exist at the same time

	datePattern string // DateTime pattern used as roll files prefix

	currentFileName string
	currentFilePath string
	currentFileSize int64
	innerWriter     io.WriteCloser // Represents file
}

// NewRollingFileWriterSize initializes a rolling writer with a 'Size' rolling mode
func NewRollingFileWriterSize(filePath string, maxFileSize int64, maxRolls int) (*RollingFileWriter, error) {
	if maxFileSize <= 0 {
		return nil, errors.New("maxFileSize must be positive")
	}

	if maxRolls <= 0 {
		return nil, errors.New("maxFileSize must be positive")
	}

	rollingFile := new(RollingFileWriter)
	rollingFile.rollingType = Size
	rollingFile.maxFileSize = maxFileSize
	rollingFile.maxRolls = maxRolls
	rollingFile.filePath = filePath
	rollingFile.fileDir, rollingFile.fileName = filepath.Split(filePath)

	return rollingFile, nil
}

// NewRollingFileWriterSize initializes a rolling writer with a 'Date' rolling mode
func NewRollingFileWriterDate(filePath string, datePattern string) (*RollingFileWriter, error) {
	rollingFile := new(RollingFileWriter)
	rollingFile.rollingType = Date
	rollingFile.datePattern = datePattern
	rollingFile.filePath = filePath
	rollingFile.fileDir, rollingFile.fileName = filepath.Split(filePath)

	return rollingFile, nil
}

func (rollFileWriter *RollingFileWriter) getFileName() string {
	if rollFileWriter.rollingType == Size {
		return rollFileWriter.fileName
	} else if rollFileWriter.rollingType == Date {
		return time.Now().Format(rollFileWriter.datePattern) + " " + rollFileWriter.fileName
	}

	return rollFileWriter.fileName
}

func (rollFileWriter *RollingFileWriter) isTimeToCreateFile() bool {
	if rollFileWriter.innerWriter == nil {
		return true
	}

	if rollFileWriter.rollingType == Size {
		return rollFileWriter.currentFileSize >= rollFileWriter.maxFileSize
	} else if rollFileWriter.rollingType == Date {
		fileName := rollFileWriter.getFileName()
		return rollFileWriter.currentFileName != fileName
	}

	return false
}

func (rollFileWriter *RollingFileWriter) createFile() error {
	if rollFileWriter.innerWriter == nil {
		return rollFileWriter.createFileAndFolderIfNeeded()
	}

	if rollFileWriter.rollingType == Size {
		if rollFileWriter.innerWriter != nil {
			rollFileWriter.innerWriter.Close()
		}

		nextRollName, err := rollFileWriter.getNextRollName()
		if err != nil {
			return err
		}

		err = fileSystemWrapper.Rename(rollFileWriter.currentFilePath, filepath.Join(rollFileWriter.fileDir, nextRollName))
		if err != nil {
			return err
		}

		rollFileWriter.deleteOldRolls()

		return rollFileWriter.createFileAndFolderIfNeeded()
	} else if rollFileWriter.rollingType == Date {
		return rollFileWriter.createFileAndFolderIfNeeded()
	}

	return nil
}

func (rollFileWriter *RollingFileWriter) getNextRollName() (string, error) {
	rolls, err := rollFileWriter.getRolls()
	if err != nil {
		return "", err
	}

	var nextIndex = 1
	for index, _ := range rolls {
		if index >= nextIndex {
			nextIndex = index + 1
		}
	}

	return rollFileWriter.currentFileName + "." + strconv.Itoa(nextIndex), nil
}

func (rollFileWriter *RollingFileWriter) getRolls() (map[int]string, error) {
	files, err := fileSystemWrapper.GetFileNames(rollFileWriter.fileDir)

	if err != nil {
		return map[int]string{}, err
	}

	rolls := make(map[int]string, 0)

	for _, file := range files {
		if strings.HasPrefix(file, rollFileWriter.currentFileName) {
			if len(rollFileWriter.currentFileName)+1 >= len(file) {
				continue
			}

			fileIndex := file[len(rollFileWriter.currentFileName)+1:]
			index, err := strconv.Atoi(fileIndex)
			if err != nil {
				continue
			}

			rolls[index] = file
		}
	}

	return rolls, nil
}

func (rollFileWriter *RollingFileWriter) deleteOldRolls() error {
	if rollFileWriter.maxRolls <= 0 {
		return nil
	}

	rolls, err := rollFileWriter.getRolls()
	if err != nil {
		return err
	}

	rollsToDelete := len(rolls) - rollFileWriter.maxRolls
	if rollsToDelete <= 0 {
		return nil
	}

	sortedRolls := rollFileWriter.sortRollsByIndex(rolls)
	for i := 0; i < rollsToDelete; i++ {
		err := fileSystemWrapper.Remove(filepath.Join(rollFileWriter.fileDir, sortedRolls[i]))
		if err != nil {
			return err
		}
	}

	return nil
}

func (rollFileWriter *RollingFileWriter) sortRollsByIndex(rolls map[int]string) []string {
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

func (rollFileWriter *RollingFileWriter) Close() error {
	return rollFileWriter.innerWriter.Close()
}

func (rollFileWriter *RollingFileWriter) Write(bytes []byte) (n int, err error) {
	if rollFileWriter.isTimeToCreateFile() {
		err := rollFileWriter.createFile()
		if err != nil {
			return 0, err
		}
	}

	if rollFileWriter.innerWriter != nil {
		rollFileWriter.currentFileSize += int64(len(bytes))
		return rollFileWriter.innerWriter.Write(bytes)
	}

	return 0, nil
}

func (rollFileWriter *RollingFileWriter) createFileAndFolderIfNeeded() error {
	err := fileSystemWrapper.MkdirAll(rollFileWriter.fileDir)
	if err != nil {
		return err
	}

	if rollFileWriter.innerWriter != nil {
		rollFileWriter.innerWriter.Close()
	}

	fileName := rollFileWriter.getFileName()
	filePath := filepath.Join(rollFileWriter.fileDir, fileName)

	var innerWriter io.WriteCloser
	if fileSystemWrapper.Exists(filePath) {
		innerWriter, err = fileSystemWrapper.Open(filePath)
		size, err := fileSystemWrapper.GetFileSize(filePath)
		if err != nil {
			return err
		}
		rollFileWriter.currentFileSize = size
	} else {
		innerWriter, err = fileSystemWrapper.Create(filePath)
		rollFileWriter.currentFileSize = 0
	}
	if err != nil {
		return err
	}

	rollFileWriter.currentFilePath = filePath
	rollFileWriter.currentFileName = fileName
	rollFileWriter.innerWriter = innerWriter

	return nil
}

func (rollFileWriter *RollingFileWriter) String() string {
	
	rollingTypeStr, ok := rollingTypesStringRepresentation[rollFileWriter.rollingType]
	if !ok {
		rollingTypeStr = "UNKNOWN"
	}

	s := fmt.Sprintf("Rolling file writer: filename: %s type: %s ", rollFileWriter.fileName, rollingTypeStr)

	if rollFileWriter.rollingType == Size {
		s += fmt.Sprintf("maxFileSize: %v, maxRolls: %v", rollFileWriter.maxFileSize, rollFileWriter.maxRolls)
	} else if rollFileWriter.rollingType == Date {
		s += fmt.Sprintf("datePattern: %v", rollFileWriter.datePattern)
	}

	return s
}
