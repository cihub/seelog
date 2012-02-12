// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package writers

// Real FileSystemWrapperInterface implementation that uses os package.

import (
	"github.com/cihub/sealog/test"
	"io"
	"os"
)

const (
	defaultFilePermissions      = 0666
	defaultDirectoryPermissions = 0767
)

var fileSystemWrapper test.FileSystemWrapperInterface = new(osWrapper)

// SetTestMode is used for testing purposes only! Do not use that or you may get incorrect behavior
func SetTestMode(testWrapper test.FileSystemWrapperInterface) {
	fileSystemWrapper = testWrapper
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
	return os.OpenFile(fileName, os.O_WRONLY | os.O_APPEND, defaultFilePermissions)
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
