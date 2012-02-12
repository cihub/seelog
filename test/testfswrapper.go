// Copyright (c) 2012 - Cloud Instruments Co. Ltd.
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

package test

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
)

// FileSystemTestWrapper emulates some of the real filesystem functions. It stores lists of
// files as if they were real files and emulate such operations as creation, folder creation, 
// renaming, removing, and others.
type FileSystemTestWrapper struct {
	root        *DirectoryWrapper
	writeCloser io.WriteCloser
	fileSize    int64
}

// NewFSTestWrapper creates a new fs wrapper for testing purposes.
func NewFSTestWrapper(root *DirectoryWrapper, writeCloser io.WriteCloser, fileSize int64) (*FileSystemTestWrapper, error) {
	if root == nil {
		var err error
		root, err = NewEmptyDirectoryWrapper("")
		if err != nil {
			return nil, err
		}
	}
	root.Name = ""
	return &FileSystemTestWrapper{root, writeCloser, fileSize}, nil
}

func NewEmptyFSTestWrapper() (*FileSystemTestWrapper, error) {
	return NewFSTestWrapper(nil, new(NullWriter), 0)
}

func (testFS *FileSystemTestWrapper) Files() []string {
	return testFS.root.GetFilesRecursively()
}

func (testFS *FileSystemTestWrapper) Exists(path string) bool {
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

func (testFS *FileSystemTestWrapper) MkdirAll(dirPath string) error {
	pathParts := strings.Split(dirPath, string(filepath.Separator))

	currentDirectory := testFS.root
	for _, pathPart := range pathParts {
		nextDirectory, found := currentDirectory.FindDirectory(pathPart)
		if !found {
			newDirectory, err := NewEmptyDirectoryWrapper(pathPart)
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

func (testFS *FileSystemTestWrapper) Open(filePath string) (io.WriteCloser, error) {
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
func (testFS *FileSystemTestWrapper) Create(filePath string) (io.WriteCloser, error) {
	directoryPath, fileName := filepath.Split(filePath)
	directory, found := testFS.root.FindDirectoryRecursively(directoryPath)
	if !found {
		return nil, errors.New("Directory not found: " + directoryPath)
	}

	if !testFS.Exists(filePath) {
		directory.Files = append(directory.Files, NewFileWrapper(fileName))
	}

	return testFS.writeCloser, nil
}
func (testFS *FileSystemTestWrapper) GetFileSize(fileName string) (int64, error) {
	return testFS.fileSize, nil
}
func (testFS *FileSystemTestWrapper) GetFileNames(folderPath string) ([]string, error) {
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
func (testFS *FileSystemTestWrapper) Rename(fileNameFrom string, fileNameTo string) error {
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
func (testFS *FileSystemTestWrapper) Remove(path string) error {
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

//=======================================================

type DirectoryWrapper struct {
	Name        string
	Directories []*DirectoryWrapper
	Files       []*FileWrapper
}

func NewDirectoryWrapper(name string, directories []*DirectoryWrapper, files []*FileWrapper) (*DirectoryWrapper, error) {
	if directories == nil {
		return nil, errors.New("directories param is nil")
	}
	if files == nil {
		return nil, errors.New("files environment param is nil")
	}

	return &DirectoryWrapper{name, directories, files}, nil
}
func NewEmptyDirectoryWrapper(name string) (*DirectoryWrapper, error) {
	return NewDirectoryWrapper(name, make([]*DirectoryWrapper, 0), make([]*FileWrapper, 0))
}
func (directory *DirectoryWrapper) GetFilesRecursively() []string {
	files := make([]string, 0)

	for _, file := range directory.Files {
		files = append(files, file.Name)
	}

	for _, directory := range directory.Directories {
		for _, fileName := range directory.GetFilesRecursively() {
			files = append(files, filepath.Join(directory.Name, fileName))
		}
	}

	return files
}
func (directory *DirectoryWrapper) FindFile(name string) (*FileWrapper, bool) {
	for _, file := range directory.Files {
		if file.Name == name {
			return file, true
		}
	}

	return nil, false
}
func (directory *DirectoryWrapper) FindDirectory(name string) (*DirectoryWrapper, bool) {
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
func (directory *DirectoryWrapper) FindDirectoryRecursively(directoryPath string) (*DirectoryWrapper, bool) {
	pathParts := strings.Split(directoryPath, string(filepath.Separator))

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

type FileWrapper struct {
	Name string
}

func NewFileWrapper(fileName string) *FileWrapper {
	return &FileWrapper{fileName}
}
