// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
