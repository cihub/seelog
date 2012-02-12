// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package writers contains a collection of writers that could be used by seelog dispatchers.
// It allows to write to such receivers as: file, console, rolling(rotation) files, smtp, network, buffered file streams.
package writers

import (
	"fmt"
	"io"
	"path/filepath"
)

// FileWriter is used to write to a file.
type FileWriter struct {
	innerWriter io.WriteCloser
	fileName    string
}

// Creates a new file and a corresponding writer. Returns error, if the file couldn't be created.
func NewFileWriter(fileName string) (writer *FileWriter, err error) {
	newWriter := new(FileWriter)

	newWriter.fileName = fileName

	fileErr := newWriter.createFile()
	if fileErr != nil {
		return nil, fileErr
	}

	return newWriter, nil
}

func (fileWriter *FileWriter) Close() error {
	return fileWriter.innerWriter.Close()
}

// Create folder and file on WriteLog/Write first call
func (fileWriter *FileWriter) Write(bytes []byte) (n int, err error) {
	return fileWriter.innerWriter.Write(bytes)
}

func (fileWriter *FileWriter) createFile() error {

	folder, _ := filepath.Split(fileWriter.fileName)

	err := fileSystemWrapper.MkdirAll(folder)

	if err != nil {
		return err
	}

	var innerWriter io.WriteCloser
	if fileSystemWrapper.Exists(fileWriter.fileName) {
		innerWriter, err = fileSystemWrapper.Open(fileWriter.fileName)
	} else {
		innerWriter, err = fileSystemWrapper.Create(fileWriter.fileName)
	}
	if err != nil {
		return err
	}

	fileWriter.innerWriter = innerWriter

	return nil
}

func (fileWriter *FileWriter) String() string {
	return fmt.Sprintf("File writer: %s", fileWriter.fileName)
}
