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

// Package seelog implements logging functionality with flexible dispatching, filtering, and formatting.
package seelog

import (
	"errors"
	"fmt"
	"time"
	"sync"
	cfg "github.com/cihub/seelog/config"
)

var Current LoggerInterface
var Default LoggerInterface
var Disabled LoggerInterface

var pkgOperationsMutex *sync.Mutex

func init() {
	pkgOperationsMutex = new(sync.Mutex)
	var err error
	
	if Default == nil {
		Default, err = LoggerFromConfigAsBytes([]byte("<seelog />"))
	}
	if Disabled == nil {
		Disabled, err = LoggerFromConfigAsBytes([]byte("<seelog levels=\"off\"/>"))
	}

	if err != nil {
		panic(fmt.Sprintf("Seelog couldn't start. Error: %s", err.Error()))
	}
	
	Current = Default
}

func createLoggerFromConfig(config *cfg.LogConfig) (LoggerInterface, error) {
	if config.LogType == cfg.SyncLoggerType {
		return NewSyncLogger(config), nil
	} else if config.LogType == cfg.AsyncLoopLoggerType {
		return NewAsyncLoopLogger(config), nil
	} else if config.LogType == cfg.AsyncTimerLoggerType {
		logData := config.LoggerData
		
		if logData == nil {
			return nil, errors.New("Invalid async timer interval!")
		}
		
		asyncInt, ok := logData.(cfg.AsyncTimerLoggerData)
		
		if !ok {
			return nil, errors.New("Invalid async timer data!")
		}
		
		return NewAsyncTimerLogger(config, time.Duration(asyncInt.AsyncInterval)), nil
	}
	return nil, errors.New("Invalid config log type/data")
}

// UseConfig uses the given logger for all Trace/Debug/... funcs.
func UseLogger(logger LoggerInterface) error {
	pkgOperationsMutex.Lock()
	if logger == nil {
		return errors.New("Logger can not be nil")
	}
	
	oldLogger := Current
	Current = logger
	
	if oldLogger != nil {
		oldLogger.Flush()
	}
	pkgOperationsMutex.Unlock()
	return nil
}

// Acts as UseLogger but the logger that was previously used would be disposed (except Default and Disabled loggers).
func ReplaceLogger(logger LoggerInterface) error {
	pkgOperationsMutex.Lock()
	if logger == nil {
		return errors.New("Logger can not be nil")
	}
	
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	if Current == Default {
		Current.Flush()
	} else if Current != nil && !Current.Closed() &&
		Current != Disabled {
			
		Current.Flush()
		Current.Close()
	} 
	
	
	
	Current = logger
	pkgOperationsMutex.Unlock()
	return nil
}

// Trace formats message according to format specifier and writes to default logger with log level = Trace
func Trace(format string, params ...interface{}) {
	pkgOperationsMutex.Lock()
	Current.Trace(format, params...)
	pkgOperationsMutex.Unlock()
}

// Debug formats message according to format specifier and writes to default logger with log level = Debug
func Debug(format string, params ...interface{}) {
	pkgOperationsMutex.Lock()
	Current.Debug(format, params...)
	pkgOperationsMutex.Unlock()
}

// Info formats message according to format specifier and writes to default logger with log level = Info
func Info(format string, params ...interface{}) {
	pkgOperationsMutex.Lock()
	Current.Info(format, params...)
	pkgOperationsMutex.Unlock()
}

// Warn formats message according to format specifier and writes to default logger with log level = Warn
func Warn(format string, params ...interface{}) {
	pkgOperationsMutex.Lock()
	Current.Warn(format, params...)
	pkgOperationsMutex.Unlock()
}

// Error formats message according to format specifier and writes to default logger with log level = Error
func Error(format string, params ...interface{}) {
	pkgOperationsMutex.Lock()
	Current.Error(format, params...)
	pkgOperationsMutex.Unlock()
}

// Critical formats message according to format specifier and writes to default logger with log level = Critical
func Critical(format string, params ...interface{}) {
	pkgOperationsMutex.Lock()
	Current.Critical(format, params...)
	pkgOperationsMutex.Unlock()
}

// Flush performs all cleanup, flushes all queued messages, etc. Call this method when your app
// is going to shut down not to lose any log messages.
func Flush() {
	pkgOperationsMutex.Lock()
	Current.Flush()
	pkgOperationsMutex.Unlock()
}
