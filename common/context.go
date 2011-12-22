// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"os"
	"errors"
	"runtime"
	"strings"
	"path/filepath"
)

// Represents runtime caller context
type LogContext struct {
	funcName  string
	shortPath string
	fullPath  string
	fileName string
}

// Returns context of the caller
func CurrentContext() (*LogContext, error) {
	return SpecificContext(1)
}

// Returns context of the function with placed "skip" stack frames of the caller
// If skip == 0 then behaves like CurrentContext
func SpecificContext(skip int) (*LogContext, error) {
	if skip < 0 {
		return nil, errors.New("Can not skip negative stack frames")
	}
	
	fullPath, shortPath, function, err := extractCallerInfo(skip + 2)
	if err != nil {
		return nil, err
	}
	_, fileName := filepath.Split(fullPath)
	return &LogContext{function, shortPath, fullPath, fileName}, nil
}

func (context *LogContext) Func() string {
	return context.funcName
}

func (context *LogContext) ShortPath() string {
	return context.shortPath
}

func (context *LogContext) FullPath() string {
	return context.fullPath
}

func (context *LogContext) FileName() string {
	return context.fileName
}

var workingDir = ""

func init() {
	setWorkDir()
}

func setWorkDir() {
	workDir, workingDirError := os.Getwd()
	if workingDirError != nil {
		workingDir = "/"
		return
	}

	workingDir = workDir + "/"
}

func extractCallerInfo(skip int) (fullPath string, shortPath string, funcName string,err error) {
	pc, fullPath, _, ok := runtime.Caller(skip)

	if !ok {
		return "", "", "", errors.New("Error during runtime.Caller")
	}

	
	if strings.HasPrefix(fullPath, workingDir) {
		shortPath = fullPath[len(workingDir):]
	} else {
		shortPath = fullPath
	}

	funName := runtime.FuncForPC(pc).Name()
	var functionName string
	if strings.HasPrefix(funName, workingDir) {
		functionName = funName[len(workingDir):len(funName)]
	} else {
		functionName = funName
	}

	return fullPath, shortPath, functionName, nil
}
