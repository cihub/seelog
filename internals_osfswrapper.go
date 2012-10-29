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
	"errors"
	"io"
	"os"
	"path/filepath"
)

const (
	defaultFilePermissions      = 0666
	defaultDirectoryPermissions = 0767
)

var isFakeFS = false
var rFSWrapper = new(realFSWrapper)
var fileSystemWrapper fileSystemWrapperInterface = rFSWrapper

func switchToRealFSWrapper() {
	if !isFakeFS {
		return
	}

	fileSystemWrapper = rFSWrapper
	isFakeFS = false
}

type realFSWrapper struct{}

func (_ *realFSWrapper) MkdirAll(folderPath string) error {
	if folderPath == "" {
		return nil
	}

	return os.MkdirAll(folderPath, defaultDirectoryPermissions)
}

func (_ *realFSWrapper) Open(fileName string) (io.WriteCloser, error) {
	return os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND, defaultFilePermissions)
}

func (_ *realFSWrapper) Create(fileName string) (io.WriteCloser, error) {
	return os.Create(fileName)
}

func (_ *realFSWrapper) GetFileSize(fileName string) (int64, error) {
	stat, err := os.Lstat(fileName)
	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}

var pathToNameTransformer filePathTransformer = func(filePath string) string {
	return filepath.Base(filePath)
}

func (_ *realFSWrapper) GetDirFileNames(dirPath string, nameIsFullPath bool) ([]string, error) {
	if dirPath == "" {
		dirPath = "."
	}

	// folder, err := os.Open(dirPath)
	// if err != nil {
	// 	return nil, err
	// }
	// defer folder.Close()

	// files, err := folder.Readdirnames(-1)
	// if err != nil {
	// 	return nil, err
	// }

	// return files, nil
	if nameIsFullPath {
		return getDirFilePaths(dirPath, nil, nil)
	}
	return getDirFilePaths(dirPath, nil, pathToNameTransformer)
}

// fileFilter accepts a file FileInfo and applies custom filtering rules.
// Returns false if file with the given FileInfo must be ignored.
type fileFilter func(os.FileInfo) bool

// filePathTransformer accepts a file path string and transforms it
// according to the custom rules.
type filePathTransformer func(string) string

func getDirFilePaths(dirPath string, fileFilter fileFilter, pathTransformer filePathTransformer) ([]string, error) {
	dfi, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dfi.Close()
	// Check if dirPath is really directory.
	if s, e := dfi.Stat(); e != nil {
		return nil, e
	} else {
		if !s.IsDir() {
			return nil, errors.New("Input path must be directory.")
		}
	}
	// Read chunck size.
	rbs := 64
	var fPaths []string
	var fp string
	var fis []os.FileInfo
	var e error
L:
	for {
		// Read directory entities by reasonable chuncks
		// to prevent overflows on big number of files.
		fis, e = dfi.Readdir(rbs)
		switch e {
		// It's OK: Do nothing, just continue the cycle.
		case nil:
		case io.EOF:
			break L
		// Something went wrong: Exit with an error.
		default:
			return nil, e
		}
		for _, fi := range fis {
			// Ignore directories.
			if !fi.IsDir() {
				// Check filter condition.
				if fileFilter != nil && !fileFilter(fi) {
					continue
				}
				fp = filepath.Join(dirPath, fi.Name())
				if pathTransformer == nil {
					fPaths = append(fPaths, fp)
				} else {
					fPaths = append(fPaths, pathTransformer(fp))
				}
			}
		}
	}
	return fPaths, nil
}

func (_ *realFSWrapper) Rename(fileNameFrom string, fileNameTo string) error {
	return os.Rename(fileNameFrom, fileNameTo)
}

func (_ *realFSWrapper) Remove(fileName string) error {
	return os.Remove(fileName)
}

func (_ *realFSWrapper) Exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
