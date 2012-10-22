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

// Real fileSystemWrapperInterface implementation that uses os package.

import (
	"io"
	"os"
)

const (
	defaultFilePermissions      = 0666
	defaultDirectoryPermissions = 0767
)

var isFakeFS = false
var realFSWrapper = new(osWrapper)
var fileSystemWrapper fileSystemWrapperInterface = realFSWrapper


func switchToRealFSWrapper() {
	if !isFakeFS {
		return
	}

	fileSystemWrapper = realFSWrapper
	isFakeFS = false
}

type osWrapper struct {
}

func (_ *osWrapper) MkdirAll(folderPath string) error {
	if folderPath == "" {
		return nil
	}

	return os.MkdirAll(folderPath, defaultDirectoryPermissions)
}
func (_ *osWrapper) Open(fileName string) (io.WriteCloser, error) {
	return os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND, defaultFilePermissions)
}
func (_ *osWrapper) Create(fileName string) (io.WriteCloser, error) {
	return os.Create(fileName)
}
func (_ *osWrapper) GetFileSize(fileName string) (int64, error) {
	stat, err := os.Lstat(fileName)
	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}
func (_ *osWrapper) GetFileNames(folderPath string) ([]string, error) {
	if folderPath == "" {
		folderPath = "."
	}

	folder, err := os.Open(folderPath)
	if err != nil {
		return make([]string, 0), err
	}
	defer folder.Close()

	files, err := folder.Readdirnames(-1)
	if err != nil {
		return make([]string, 0), err
	}

	return files, nil
}
func (_ *osWrapper) Rename(fileNameFrom string, fileNameTo string) error {
	return os.Rename(fileNameFrom, fileNameTo)
}
func (_ *osWrapper) Remove(fileName string) error {
	return os.Remove(fileName)
}
func (_ *osWrapper) Exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
