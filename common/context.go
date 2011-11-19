package common

import (
	"os"
	"runtime"
	"strings"
)

// Represents runtime caller context
type LogContext struct {
	funcName  string
	shortPath string
	fullPath  string
}

func CurrentContext() (*LogContext, os.Error) {
	fullPath, shortPath, function, err := extractCallerInfo(2)
	if err != nil {
		return nil, err
	}

	return &LogContext{function, shortPath, fullPath}, nil
}

func (this *LogContext) Func() string {
	return this.funcName
}

func (this *LogContext) ShortPath() string {
	return this.shortPath
}

func (this *LogContext) FullPath() string {
	return this.fullPath
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

func extractCallerInfo(skip int) (string, string, string, os.Error) {
	pc, fullPath, _, ok := runtime.Caller(skip)

	if !ok {
		return "", "", "", os.NewError("Error during runtime.Caller")
	}

	var shortPath string
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
