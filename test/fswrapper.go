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
	"testing"
)

// FileSystemWrapperInterface is used for testing. When seelog is used in a real app, osWrapper uses standard os
// funcs. When seelog is being tested, FileSystemTestWrapper emulates some of the os funcs. Both osWrapper and
// FileSystemTestWrapper implement this interface.
type FileSystemWrapperInterface interface {
	MkdirAll(folderPath string) error
	Open(fileName string) (io.WriteCloser, error)
	Create(fileName string) (io.WriteCloser, error)
	GetFileSize(fileName string) (int64, error)
	GetFileNames(folderPath string) ([]string, error)
	Rename(fileNameFrom string, fileNameTo string) error
	Remove(fileName string) error
	Exists(path string) bool
}

var wrapperForTest FileSystemWrapperInterface

func SetWrapperTestEnvironment(wrapper FileSystemWrapperInterface) {
	wrapperForTest = wrapper
}

func removeAndCheck(fileName string) error {
	err := wrapperForTest.Remove(fileName)
	if err != nil {
		return err
	}

	if wrapperForTest.Exists(fileName) {
		return errors.New("Must be deleted: " + fileName)
	}

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

func TestFswrapper_RemoveFile(t *testing.T) {
	fileName := "file.txt"

	err := createFile(fileName)
	if err != nil {
		t.Error(err)
		return
	}

	err = removeAndCheck(fileName)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestFswrapper_RemoveFolder(t *testing.T) {
	folderName := "testFolder"

	err := wrapperForTest.MkdirAll(folderName)
	if err != nil {
		t.Error(err)
		return
	}

	err = removeAndCheck(folderName)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestFswrapper_MkdirAll(t *testing.T) {
	parentFolder := "testWrapper"
	nestedFolder := filepath.Join(parentFolder, "test")

	err := wrapperForTest.MkdirAll(nestedFolder)
	if err != nil {
		t.Error(err)
		return
	}

	if !wrapperForTest.Exists(nestedFolder) {
		t.Error("Expected folder: " + nestedFolder)
		return
	}

	err = removeAndCheck(nestedFolder)
	if err != nil {
		t.Error(err)
		return
	}

	err = removeAndCheck(parentFolder)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestFswrapper_CreateNewFile(t *testing.T) {
	fileName := "file.txt"

	err := createFile(fileName)
	if err != nil {
		t.Error(err)
		return
	}

	if !wrapperForTest.Exists(fileName) {
		t.Error("Expected file: " + fileName)
		return
	}

	err = removeAndCheck(fileName)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestFswrapper_OpenFile(t *testing.T) {
	fileName := "file.txt"

	err := createFile(fileName)
	if err != nil {
		t.Error(err)
		return
	}

	err = createFile(fileName)
	if err != nil {
		t.Error(err)
		return
	}

	if !wrapperForTest.Exists(fileName) {
		t.Error("Expected file: " + fileName)
		return
	}

	err = removeAndCheck(fileName)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestFswrapper_GetFileSize(t *testing.T) {
	fileName := "file.txt"
	data := []byte("hello")

	file, err := wrapperForTest.Create(fileName)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = file.Write(data)
	if err != nil {
		t.Error(err)
		return
	}

	err = file.Close()
	if err != nil {
		t.Error(err)
		return
	}

	size, err := wrapperForTest.GetFileSize(fileName)
	if err != nil {
		t.Error(err)
		return
	}

	if int64(len(data)) != size {
		t.Errorf("Incorrect file size. Expected:%v, got:%v", len(data), size)
		return
	}

	err = removeAndCheck(fileName)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestFswrapper_Rename(t *testing.T) {
	fileNameFrom := "file.txt"
	fileNameTo := "file1.txt"

	err := createFile(fileNameFrom)
	if err != nil {
		t.Error(err)
		return
	}

	err = wrapperForTest.Rename(fileNameFrom, fileNameTo)
	if err != nil {
		t.Error(err)
		return
	}

	if wrapperForTest.Exists(fileNameFrom) {
		t.Error("File must be deleted: " + fileNameFrom)
		return
	}
	if !wrapperForTest.Exists(fileNameTo) {
		t.Error("Missing file: " + fileNameTo)
		return
	}

	err = removeAndCheck(fileNameTo)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestFswrapper_GetFileNames(t *testing.T) {
	folder := "testWrapper"
	files := []string{"file1", "file2", "file3"}

	err := wrapperForTest.MkdirAll(folder)
	if err != nil {
		t.Error(err)
		return
	}

	for _, fileName := range files {
		err = createFile(filepath.Join(folder, fileName))
		if err != nil {
			t.Error(err)
			return
		}
	}

	filesFromWrapper, err := wrapperForTest.GetFileNames(folder)
	if err != nil {
		t.Error(err)
		return
	}

	for _, file := range files {
		exists := false
		for _, fileFromWrapper := range filesFromWrapper {
			if file == fileFromWrapper {
				exists = true
				break
			}
		}
		if !exists {
			t.Error("Missing: '" + file + "'")
		}
	}

	for _, fileFromWrapper := range filesFromWrapper {
		exists := false
		for _, file := range files {
			if file == fileFromWrapper {
				exists = true
				break
			}
		}
		if !exists {
			t.Error("Excess: '" + fileFromWrapper + "'")
		}
	}

	for _, fileName := range files {
		err = removeAndCheck(filepath.Join(folder, fileName))
		if err != nil {
			t.Error(err)
			return
		}
	}

	err = removeAndCheck(folder)
	if err != nil {
		t.Error(err)
		return
	}
}
