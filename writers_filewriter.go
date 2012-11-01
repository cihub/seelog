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
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// fileWriter is used to write to a file.
type fileWriter struct {
	innerWriter io.WriteCloser
	fileName    string
}

// Creates a new file and a corresponding writer. Returns error, if the file couldn't be created.
func newFileWriter(fileName string) (writer *fileWriter, err error) {
	newWriter := new(fileWriter)

	newWriter.fileName = fileName

	fileErr := newWriter.createFile()
	if fileErr != nil {
		return nil, fileErr
	}

	return newWriter, nil
}

func (fileWriter *fileWriter) Close() error {
	return fileWriter.innerWriter.Close()
}

// Create folder and file on WriteLog/Write first call
func (fileWriter *fileWriter) Write(bytes []byte) (n int, err error) {
	return fileWriter.innerWriter.Write(bytes)
}

func (fileWriter *fileWriter) createFile() error {

	folder, _ := filepath.Split(fileWriter.fileName)
	var err error

	if 0 != len(folder) {
		err = os.MkdirAll(folder, defaultDirectoryPermissions)

		if err != nil {
			return err
		}
	}

	// If exists
	_, err = os.Lstat(fileWriter.fileName)
	if nil == err {
		fileWriter.innerWriter, err = os.OpenFile(fileWriter.fileName, os.O_WRONLY|os.O_APPEND, defaultFilePermissions)
	} else {
		fileWriter.innerWriter, err = os.Create(fileWriter.fileName)
	}
	if err != nil {
		return err
	}

	return nil
}

func (fileWriter *fileWriter) String() string {
	return fmt.Sprintf("File writer: %s", fileWriter.fileName)
}
