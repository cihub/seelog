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
	"io"
)

// fileSystemWrapperInterface is designed to allow for flexible testing.
// When seelog is used in a real app, realFSWrapper uses the standard os package funcs,
// when tested - testFileSystemWrapper just emulates the respective os package funcs.
// Both realFSWrapper and testFileSystemWrapper implement this interface.
type fileSystemWrapperInterface interface {
	MkdirAll(folderPath string) error
	Open(fileName string) (io.WriteCloser, error)
	Create(fileName string) (io.WriteCloser, error)
	GetFileSize(fileName string) (int64, error)
	GetDirFileNames(dirPath string, nameIsFullPath bool) ([]string, error)
	Rename(fileNameFrom string, fileNameTo string) error
	Remove(fileName string) error
	Exists(path string) bool
}

var wrapperForTest fileSystemWrapperInterface

func setWrapperTestEnvironment(fsWrapper fileSystemWrapperInterface) {
	wrapperForTest = fsWrapper
}

func removeAndCheck(fileName string) error {
	err := wrapperForTest.Remove(fileName)
	if err != nil {
		return err
	}

	// if wrapperForTest.Exists(fileName) {
	// 	return errors.New("Must be deleted: " + fileName)
	// }

	return nil
}

func createFile(fileName string) error {
	file, err := wrapperForTest.Create(fileName)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}
