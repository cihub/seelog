// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sealog implements logging functionality with flexible dispatching, filtering, and formatting.
package sealog

import (
	"errors"
	"fmt"
	"time"
	cfg "github.com/cihub/sealog/config"
)

var currentLogger LoggerInterface
var defaultConfig *cfg.LogConfig

func init() {
	if defaultConfig == nil {
		var err error
		defaultConfig, err = ConfigFromBytes([]byte("<sealog />"))
		if err != nil {
			panic(fmt.Sprintf("Sealog couldn't start. Error: %s", err.Error()))
		}
	}
	UseDefaultConfig()
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

// UseConfig uses the given configuration to create a logger from it and use it
// for all Trace/Debug/... funcs.
// The logger that was previously used would be disposed.
func UseConfig(config *cfg.LogConfig) error {
	if config == nil {
		return errors.New("Config can not be nil")
	}
	
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	if currentLogger != nil && !currentLogger.Closed() {
		currentLogger.Flush()
		currentLogger.Close()
	}
	
	newLogger, err := createLoggerFromConfig(config)
	
	if err == nil {
		currentLogger = newLogger
	}

	return err
}

// UseDefaultConfig uses default configuration
func UseDefaultConfig() error {
	return UseConfig(defaultConfig)
}

// Trace formats message according to format specifier and writes to default logger with log level = Trace
func Trace(format string, params ...interface{}) {
	currentLogger.Trace(format, params...)
}

// Debug formats message according to format specifier and writes to default logger with log level = Debug
func Debug(format string, params ...interface{}) {
	currentLogger.Debug(format, params...)
}

// Info formats message according to format specifier and writes to default logger with log level = Info
func Info(format string, params ...interface{}) {
	currentLogger.Info(format, params...)
}

// Warn formats message according to format specifier and writes to default logger with log level = Warn
func Warn(format string, params ...interface{}) {
	currentLogger.Warn(format, params...)
}

// Error formats message according to format specifier and writes to default logger with log level = Error
func Error(format string, params ...interface{}) {
	currentLogger.Error(format, params...)
}

// Critical formats message according to format specifier and writes to default logger with log level = Critical
func Critical(format string, params ...interface{}) {
	currentLogger.Critical(format, params...)
}

// Flush performs all cleanup, flushes all queued messages, etc. Call this method when your app
// is going to shut down not to lose any log messages.
func Flush() {
	currentLogger.Flush()
}