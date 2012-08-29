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
)

const (
	staticFuncCallDepth = 3 // See 'commonLogger.log' method comments
	loggerFuncCallDepth = 3
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

func createLoggerFromConfig(config *logConfig) (LoggerInterface, error) {
	if config.LogType == syncloggerTypeFromString {
		return newSyncLogger(config), nil
	} else if config.LogType == asyncLooploggerTypeFromString {
		return newAsyncLoopLogger(config), nil
	} else if config.LogType == asyncTimerloggerTypeFromString {
		logData := config.LoggerData
		
		if logData == nil {
			return nil, errors.New("Async timer data not set!")
		}
		
		asyncInt, ok := logData.(asyncTimerLoggerData)
		
		if !ok {
			return nil, errors.New("Invalid async timer data!")
		}
		
		logger, err := newAsyncTimerLogger(config, time.Duration(asyncInt.AsyncInterval))
		
		if !ok {
			return nil, err
		}
		
		return logger, nil
	} else if config.LogType == adaptiveLoggerTypeFromString {
		logData := config.LoggerData
		
		if logData == nil {
			return nil, errors.New("Adaptive logger parameters not set!")
		}
		
		adaptData, ok := logData.(adaptiveLoggerData)
		
		if !ok {
			return nil, errors.New("Invalid adaptive logger parameters!")
		}
		
		logger, err := newAsyncAdaptiveLogger(config, time.Duration(adaptData.MinInterval),
													  time.Duration(adaptData.MaxInterval),
													  adaptData.CriticalMsgCount)
		
		if !ok {
			return nil, err
		}
		
		return logger, nil
	}
	return nil, errors.New("Invalid config log type/data")
}

// UseConfig uses the given logger for all Trace/Debug/... funcs.
func UseLogger(logger LoggerInterface) error {
	if logger == nil {
		return errors.New("Logger can not be nil")
	}

	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	
	oldLogger := Current
	Current = logger
	
	if oldLogger != nil {
		oldLogger.Flush()
	}
	
	return nil
}

// Acts as UseLogger but the logger that was previously used would be disposed (except Default and Disabled loggers).
func ReplaceLogger(logger LoggerInterface) error {
	if logger == nil {
		return errors.New("Logger can not be nil")
	}

	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	
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
	
	return nil
}

// Tracef formats message according to format specifier and writes to default logger with log level = Trace
func Tracef(format string, params ...interface{}) {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.traceWithCallDepth(staticFuncCallDepth, newFormattedLogMessage(format, params))
}

// Debugf formats message according to format specifier and writes to default logger with log level = Debug
func Debugf(format string, params ...interface{}) {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.debugWithCallDepth(staticFuncCallDepth, newFormattedLogMessage(format, params))
}

// Infof formats message according to format specifier and writes to default logger with log level = Info
func Infof(format string, params ...interface{}) {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.infoWithCallDepth(staticFuncCallDepth, newFormattedLogMessage(format, params))
}

// Warnf formats message according to format specifier and writes to default logger with log level = Warn
func Warnf(format string, params ...interface{}) {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.warnWithCallDepth(staticFuncCallDepth, newFormattedLogMessage(format, params))
}

// Errorf formats message according to format specifier and writes to default logger with log level = Error
func Errorf(format string, params ...interface{}) {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.errorWithCallDepth(staticFuncCallDepth, newFormattedLogMessage(format, params))
}

// Criticalf formats message according to format specifier and writes to default logger with log level = Critical
func Criticalf(format string, params ...interface{}) {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.criticalWithCallDepth(staticFuncCallDepth, newFormattedLogMessage(format, params))
}

// Trace formats message using the default formats for its operands and writes to default logger with log level = Trace
func Trace(v ...interface{}) {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.traceWithCallDepth(staticFuncCallDepth, newLogMessage(v))
}

// Debug formats message using the default formats for its operands and writes to default logger with log level = Debug
func Debug(v ...interface{}) {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.debugWithCallDepth(staticFuncCallDepth, newLogMessage(v))
}

// Info formats message using the default formats for its operands and writes to default logger with log level = Info
func Info(v ...interface{}) {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.infoWithCallDepth(staticFuncCallDepth, newLogMessage(v))
}

// Warn formats message using the default formats for its operands and writes to default logger with log level = Warn
func Warn(v ...interface{}) {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.warnWithCallDepth(staticFuncCallDepth, newLogMessage(v))
}

// Error formats message using the default formats for its operands and writes to default logger with log level = Error
func Error(v ...interface{}) {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.errorWithCallDepth(staticFuncCallDepth, newLogMessage(v))
}

// Critical formats message using the default formats for its operands and writes to default logger with log level = Critical
func Critical(v ...interface{}) {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.criticalWithCallDepth(staticFuncCallDepth, newLogMessage(v))
}

// Flush performs all cleanup, flushes all queued messages, etc. Call this method when your app
// is going to shut down not to lose any log messages.
func Flush() {
	pkgOperationsMutex.Lock()
	defer pkgOperationsMutex.Unlock()
	Current.Flush()
}
