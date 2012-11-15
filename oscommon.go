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
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	defaultFilePermissions      = 0666
	defaultDirectoryPermissions = 0767
)

// fileFilter accepts a file FileInfo and applies custom filtering rules.
// Returns false if file with the given FileInfo must be ignored.
type fileFilter func(os.FileInfo) bool

// filePathTransformer accepts a file path string and transforms it
// according to the custom rules.
type filePathTransformer func(string) string

var pathToNameTransformer filePathTransformer = func(filePath string) string {
	return filepath.Base(filePath)
}

func getDirFileNames(dirPath string, nameIsFullPath bool, filter fileFilter) ([]string, error) {
	if 0 == len(dirPath) {
		dirPath = "."
	}

	// return files, nil
	if nameIsFullPath {
		return getDirFilePaths(dirPath, filter, nil)
	}
	return getDirFilePaths(dirPath, filter, pathToNameTransformer)
}

func getDirFilePaths(
	path string,
	fileFilter fileFilter,
	pathTransformer filePathTransformer) ([]string, error) {

	fis, err := ioutil.ReadDir(path)

	if nil != err {
		return nil, err
	}

	fPaths := make([]string, 0)

	for _, fi := range fis {
		// Ignore directories.
		if !fi.IsDir() {
			// Check filter condition.
			if fileFilter != nil && !fileFilter(fi) {
				continue
			}

			if pathTransformer == nil {
				fPaths = append(fPaths, fi.Name())
			} else {
				fPaths = append(fPaths, pathTransformer(fi.Name()))
			}
		}
	}

	return fPaths, nil
}

func tryRemoveFile(filePath string) error {
	err := os.Remove(filePath)
	if os.IsNotExist(err) {
		return nil
	}

	return err
}

// FileExists return flag whether a given file exists
// and operation error if an unclassified failure occurs.
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}
