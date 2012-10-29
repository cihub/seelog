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
	"io"
	"path/filepath"
	"strings"
	"testing"
)

var fakeFSWrapper fileSystemWrapperInterface

func switchToFakeFSWrapper(t *testing.T) {
	if isFakeFS {
		return
	}

	if fakeFSWrapper == nil {
		newTestFSWrapper, err := newEmptyFSTestWrapper()
		if err != nil {
			t.Fatalf("Fatal error in test fs initialization: %s", err.Error())
		}

		fakeFSWrapper = newTestFSWrapper
	}

	fileSystemWrapper = fakeFSWrapper
	isFakeFS = true
}

// testFileSystemWrapper emulates some of the real file system functions. It stores lists of
// files as if they were real files and emulate such operations as creation, folder creation, 
// renaming, removing, and others.
type testFileSystemWrapper struct {
	root        *directoryWrapper
	writeCloser io.WriteCloser
	fileSize    int64
}

// newTestFSWrapper creates a new fs wrapper for testing purposes.
func newTestFSWrapper(root *directoryWrapper, writeCloser io.WriteCloser, fileSize int64) (*testFileSystemWrapper, error) {
	if root == nil {
		var err error
		root, err = newEmptyDirectoryWrapper("")
		if err != nil {
			return nil, err
		}
	}
	root.Name = ""
	return &testFileSystemWrapper{root, writeCloser, fileSize}, nil
}

func newEmptyFSTestWrapper() (*testFileSystemWrapper, error) {
	return newTestFSWrapper(nil, new(nullWriter), 0)
}

func (testFS *testFileSystemWrapper) Files() []string {
	return testFS.root.GetFileNamesRecursively()
}

func (testFS *testFileSystemWrapper) Exists(path string) bool {
	parentDirPath, fileName := filepath.Split(path)
	parentDir, found := testFS.root.FindDirectoryRecursively(parentDirPath)
	if !found {
		return false
	}

	_, found = parentDir.FindDirectory(fileName)
	if found {
		return true
	}

	_, found = parentDir.FindFile(fileName)
	return found
}

func (testFS *testFileSystemWrapper) MkdirAll(dirPath string) error {
	pathParts := strings.Split(dirPath, string(filepath.Separator))

	currentDirectory := testFS.root
	for _, pathPart := range pathParts {
		nextDirectory, found := currentDirectory.FindDirectory(pathPart)
		if !found {
			newDirectory, err := newEmptyDirectoryWrapper(pathPart)
			if err != nil {
				return err
			}
			currentDirectory.Directories = append(currentDirectory.Directories, newDirectory)
			nextDirectory = newDirectory
		}

		currentDirectory = nextDirectory
	}

	return nil
}

func (testFS *testFileSystemWrapper) Open(filePath string) (io.WriteCloser, error) {
	directoryPath, _ := filepath.Split(filePath)
	_, found := testFS.root.FindDirectoryRecursively(directoryPath)
	if !found {
		return nil, errors.New("Directory not found: " + directoryPath)
	}

	if !testFS.Exists(filePath) {
		return nil, errors.New("File already exists " + filePath)
	}

	return testFS.writeCloser, nil
}

func (testFS *testFileSystemWrapper) Create(filePath string) (io.WriteCloser, error) {
	directoryPath, fileName := filepath.Split(filePath)
	directory, found := testFS.root.FindDirectoryRecursively(directoryPath)
	if !found {
		return nil, errors.New("Directory not found: " + directoryPath)
	}

	if !testFS.Exists(filePath) {
		directory.Files = append(directory.Files, newFileWrapper(fileName))
	}

	return testFS.writeCloser, nil
}

func (testFS *testFileSystemWrapper) GetFileSize(fileName string) (int64, error) {
	return testFS.fileSize, nil
}

func (testFS *testFileSystemWrapper) GetDirFileNames(folderPath string, nameIsFullPath bool) ([]string, error) {
	// Find given directory down to the directory tree.
	directory, found := testFS.root.FindDirectoryRecursively(folderPath)
	if !found {
		return nil, errors.New("Directory not found: " + folderPath)
	}

	files := make([]string, 0)
	for _, file := range directory.Files {
		files = append(files, file.Name)
	}

	return files, nil
}

func (testFS *testFileSystemWrapper) Rename(fileNameFrom string, fileNameTo string) error {
	if testFS.Exists(fileNameTo) {
		return errors.New("Such file already exists")
	}
	if !testFS.Exists(fileNameFrom) {
		return errors.New("Cannot rename nonexistent file")
	}

	testFS.Remove(fileNameFrom)
	testFS.Create(fileNameTo)

	return nil
}

func (testFS *testFileSystemWrapper) Remove(path string) error {
	parentDirPath, fileName := filepath.Split(path)
	parentDir, found := testFS.root.FindDirectoryRecursively(parentDirPath)
	if !found {
		return errors.New("Directory not found: " + parentDirPath)
	}

	_, found = parentDir.FindDirectory(fileName)
	if found {
		for i, dir := range parentDir.Directories {
			if dir.Name == fileName {
				parentDir.Directories = append(parentDir.Directories[:i], parentDir.Directories[i+1:]...)
				return nil
			}
		}
	}

	_, found = parentDir.FindFile(fileName)
	if found {
		for i, file := range parentDir.Files {
			if file.Name == fileName {
				parentDir.Files = append(parentDir.Files[:i], parentDir.Files[i+1:]...)
				return nil
			}
		}
	}

	return errors.New("Cannot remove nonexistent file")
}

// Wrappers for FS entities.

type fileWrapper struct {
	Name string
}

func newFileWrapper(fileName string) *fileWrapper {
	return &fileWrapper{fileName}
}

type directoryWrapper struct {
	Name        string
	Directories []*directoryWrapper
	Files       []*fileWrapper
}

func newDirectoryWrapper(name string, directories []*directoryWrapper, files []*fileWrapper) (*directoryWrapper, error) {
	if directories == nil {
		return nil, errors.New("directories param is nil")
	}
	if files == nil {
		return nil, errors.New("files environment param is nil")
	}

	return &directoryWrapper{name, directories, files}, nil
}

func newEmptyDirectoryWrapper(name string) (*directoryWrapper, error) {
	return newDirectoryWrapper(name, make([]*directoryWrapper, 0), make([]*fileWrapper, 0))
}

func (directory *directoryWrapper) GetFileNamesRecursively() []string {
	files := make([]string, 0)
	for _, file := range directory.Files {
		files = append(files, file.Name)
	}

	for _, directory := range directory.Directories {
		for _, fileName := range directory.GetFileNamesRecursively() {
			files = append(files, filepath.Join(directory.Name, fileName))
		}
	}

	return files
}

func (directory *directoryWrapper) FindFile(name string) (*fileWrapper, bool) {
	for _, file := range directory.Files {
		if file.Name == name {
			return file, true
		}
	}

	return nil, false
}

func (directory *directoryWrapper) FindDirectory(name string) (*directoryWrapper, bool) {
	if name == "" || name == "." {
		return directory, true
	}

	for _, directory := range directory.Directories {
		if directory.Name == name {
			return directory, true
		}
	}

	return nil, false
}

func (directory *directoryWrapper) FindDirectoryRecursively(dirPath string) (*directoryWrapper, bool) {
	pathParts := strings.Split(dirPath, string(filepath.Separator))
	currentDirectory := directory
	for _, pathPart := range pathParts {
		nextDirectory, found := currentDirectory.FindDirectory(pathPart)
		if !found {
			return nil, false
		}

		currentDirectory = nextDirectory
	}

	return currentDirectory, true
}
