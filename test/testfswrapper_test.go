// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"testing"
)

func init() {
	wrapper, err := NewFSTestWrapper(nil, new(NullWriter), 5)
	if err != nil {
		panic(err)
	}

	SetWrapperTestEnvironment(wrapper)
}

func TestFileSystemWrapper_MkdirAll(t *testing.T) {
	TestFswrapper_MkdirAll(t)
}

func TestFileSystemWrapper_CreateNewFile(t *testing.T) {
	TestFswrapper_CreateNewFile(t)
}

func TestFileSystemWrapper_OpenFile(t *testing.T) {
	TestFswrapper_OpenFile(t)
}

func TestFileSystemWrapper_GetFileSize(t *testing.T) {
	TestFswrapper_GetFileSize(t)
}

func TestFileSystemWrapper_GetFileNames(t *testing.T) {
	TestFswrapper_GetFileNames(t)
}

func TestFileSystemWrapper_RemoveFile(t *testing.T) {
	TestFswrapper_RemoveFile(t)
}

func TestFileSystemWrapper_RemoveFolder(t *testing.T) {
	TestFswrapper_RemoveFolder(t)
}

func TestFileSystemWrapper_Rename(t *testing.T) {
	TestFswrapper_Rename(t)
}
