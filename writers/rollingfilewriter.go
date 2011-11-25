package writers

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"strconv"
	"time"
	"sort"
)

type rollingTypes uint8

const (
	Size = iota
	Date
)

// Rolls log files by size or date.
// 1. Size mode
//   Creates log file with name from fileName parameter.
//   If the current file size exceeds maxFileSize then current file becomes a roll.
//   If the rolls count exceeds maxRolls then the older rolls deleted.
//   If maxRolls <= 0, the rolls never deleted.
// 2. Date mode
//   Creates logfile with name as [date formatted with datePattern] + fileName.
//   If due to change in time changed file name ( generated from the datePattern ) then new file created.
type RollingFileWriter struct {
	fileName    string
	folderName  string       // Rolling files folder. Set to "" to use current directory
	rollingType rollingTypes // Rolling mode (Files roll by size/date/...)

	maxFileSize int64 // Maximal file size at which roll must occur
	maxRolls    int   // Maximal count of roll files that exist at the same time

	datePattern string // DateTime pattern used as roll files prefix

	currentFileName string
	innerWriter     io.WriteCloser // Represents file
}

// Initializes writer by Size mode
func NewRollingFileWriterSize(fileName string, maxFileSize int64, maxRolls int) *RollingFileWriter {
	rollingFile := new(RollingFileWriter)
	rollingFile.rollingType = Size
	rollingFile.maxFileSize = maxFileSize
	rollingFile.maxRolls = maxRolls
	rollingFile.folderName, rollingFile.fileName = filepath.Split(fileName)
	
	return rollingFile
}

// Initializes writer by Date mode
func NewRollingFileWriterDate(fileName string, datePattern string) *RollingFileWriter {
	rollingFile := new(RollingFileWriter)
	rollingFile.rollingType = Date
	rollingFile.datePattern = datePattern
	rollingFile.folderName, rollingFile.fileName = filepath.Split(fileName)

	return rollingFile
}

func (this *RollingFileWriter) getFileName() string {
	if this.rollingType == Size {
		return filepath.Join(this.folderName, this.fileName)
	} else if this.rollingType == Date {
		return filepath.Join(this.folderName, time.LocalTime().Format(this.datePattern)+" "+this.fileName)
	}

	return filepath.Join(this.folderName, this.fileName)
}

func (this *RollingFileWriter) isTimeToCreateFile() bool {
	if this.innerWriter == nil {
		return true
	}

	if this.rollingType == Size {
		size, err := fileSystemWrapper.GetFileSize(this.currentFileName)
		if err == nil {
			if size >= this.maxFileSize {
				return true
			}
		}
	} else if this.rollingType == Date {
		return this.currentFileName != this.getFileName()
	}

	return false
}

func (this *RollingFileWriter) createFile() os.Error {
	if this.innerWriter == nil {
		return this.createFileAndFolderIfNeeded()
	}

	if this.rollingType == Size {
		if this.innerWriter != nil {
			this.innerWriter.Close()
		}

		nextRollName, err := this.getNextRollName()
		if err != nil {
			return err
		}

		err = fileSystemWrapper.Rename(this.currentFileName, nextRollName)
		if err != nil {
			return err
		}

		this.deleteOldRolls()

		return this.createFileAndFolderIfNeeded()
	} else if this.rollingType == Date {
		return this.createFileAndFolderIfNeeded()
	}

	return nil
}

func (this *RollingFileWriter) getNextRollName() (string, os.Error) {
	rolls, err := this.getRolls()
	if err != nil {
		return "", err
	}

	var nextIndex = 1
	for _, file := range rolls {
		index := this.gerRollIndex(file)

		if index >= nextIndex {
			nextIndex = index + 1
		}
	}

	return this.currentFileName + "." + strconv.Itoa(nextIndex), nil
}

// Returns -1 in case of any error
func (this *RollingFileWriter) gerRollIndex(file string) int {
	fileIndex := file[len(this.currentFileName)+1:]
	index, err := strconv.Atoi(fileIndex)
	if err != nil {
		return -1
	}

	return index
}

func (this *RollingFileWriter) getRolls() ([]string, os.Error) {
	files, err := fileSystemWrapper.GetFileNames(this.folderName)
	if err != nil {
		return []string{}, err
	}

	rolls := make([]string, 0)

	for _, file := range files {
		if strings.HasPrefix(file, this.currentFileName) {
			if len(this.currentFileName)+1 >= len(file) {
				continue
			}

			rolls = append(rolls, file)
		}
	}

	return rolls, nil
}

func (this *RollingFileWriter) deleteOldRolls() os.Error {
	if this.maxRolls <= 0 {
		return nil
	}

	rolls, err := this.getRolls()
	if err != nil {
		return nil
	}

	rollsToDelete := len(rolls) - this.maxRolls

	if rollsToDelete <= 0 {
		return nil
	}

	sortedRolls := this.sortRollsByIndex(rolls)
	for i := 0; i < rollsToDelete; i++ {
		fileSystemWrapper.Remove(sortedRolls[i])
	}

	return nil
}

func (this *RollingFileWriter) sortRollsByIndex(rolls []string) []string {
	if len(rolls) < 2 {
		return rolls
	}

	rollIndexesByIndex := make(map[int]int, 0)
	indexes := make([]int, 0)

	for rollIndex, roll := range rolls {
		index := this.gerRollIndex(roll)
		if index < 0 {
			continue
		}

		rollIndexesByIndex[index] = rollIndex
		indexes = append(indexes, index)
	}

	sort.Ints(indexes)

	sortedRolls := make([]string, len(indexes))
	for i, index := range indexes {
		sortedRolls[i] = rolls[rollIndexesByIndex[index]]
	}
	return sortedRolls
}

func (this *RollingFileWriter) Write(bytes []byte) (n int, err os.Error) {
	if this.isTimeToCreateFile() {
		err := this.createFile()
		if err != nil {
			return 0, err
		}
	}

	if this.innerWriter != nil {
		return this.innerWriter.Write(bytes)
	}

	return 0, nil
}

func (this *RollingFileWriter) createFileAndFolderIfNeeded() os.Error {
	dirErr := fileSystemWrapper.MkdirAll(this.folderName)
	if dirErr != nil {
		return dirErr
	}

	if this.innerWriter != nil {
		this.innerWriter.Close()
	}

	this.currentFileName = this.getFileName()
	var fileError os.Error
	this.innerWriter, fileError = fileSystemWrapper.Create(this.currentFileName)
	if fileError != nil {
		return fileError
	}

	return nil
}
