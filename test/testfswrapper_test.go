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
