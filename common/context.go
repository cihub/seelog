package common

import (
	"os"
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

func CurrentContext() (*LogContext, os.Error) {
	fullPath, shortPath, function, err := extractCallerInfo(2)
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

func extractCallerInfo(skip int) (fullPath string, shortPath string, funcName string,err os.Error) {
	pc, fullPath, _, ok := runtime.Caller(skip)

	if !ok {
		return "", "", "", os.NewError("Error during runtime.Caller")
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
