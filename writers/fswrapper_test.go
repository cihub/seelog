// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package writers

import (
	"testing"
	"github.com/cihub/sealog/test"
)

func init() {
	test.SetWrapperTestEnvironment(new(osWrapper))
}

func TestFileSystemWrapper_MkdirAll(t *testing.T) {
	test.TestFswrapper_MkdirAll(t)
}

func TestFileSystemWrapper_CreateNewFile(t *testing.T) {
	test.TestFswrapper_CreateNewFile(t)
}

func TestFileSystemWrapper_OpenFile(t *testing.T) {
	test.TestFswrapper_OpenFile(t)
}

func TestFileSystemWrapper_GetFileSize(t *testing.T) {
	test.TestFswrapper_GetFileSize(t)
}

func TestFileSystemWrapper_GetFileNames(t *testing.T) {
	test.TestFswrapper_GetFileNames(t)
}

func TestFileSystemWrapper_RemoveFile(t *testing.T) {
	test.TestFswrapper_RemoveFile(t)
}

func TestFileSystemWrapper_RemoveFolder(t *testing.T) {
	test.TestFswrapper_RemoveFolder(t)
}

func TestFileSystemWrapper_Rename(t *testing.T) {
	test.TestFswrapper_Rename(t)
}
