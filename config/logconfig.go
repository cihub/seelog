// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"github.com/cihub/seelog/dispatchers"
	. "github.com/cihub/seelog/common"
	"errors"
)

type LoggerType uint8

const (
	SyncLoggerType = iota
	AsyncLoopLoggerType
	AsyncTimerLoggerType
	DefaultLoggerType = AsyncLoopLoggerType
)

const (
	SyncLoggerTypeStr = "sync"
	AsyncLoggerTypeStr = "asyncloop"
	AsyncTimerLoggerTypeStr = "asynctimer"
)

// AsyncTimerLoggerData represents specific data for async timer logger
type AsyncTimerLoggerData struct {
	AsyncInterval uint32
}

var loggerTypeToStringRepresentations = map[LoggerType]string{
	SyncLoggerType:    		SyncLoggerTypeStr,
	AsyncLoopLoggerType:    AsyncLoggerTypeStr,
	AsyncTimerLoggerType:   AsyncTimerLoggerTypeStr,
}

// LoggerTypeFromString parses a string and returns a corresponding logger type, if sucessfull. 
func LoggerTypeFromString(logTypeString string) (level LoggerType, found bool) {
	for logType, logTypeStr := range loggerTypeToStringRepresentations {
		if logTypeStr == logTypeString {
			return logType, true
		}
	}

	return 0, false
}

// LogConfig stores logging configuration. Contains messages dispatcher, allowed log level rules 
// (general constraints and exceptions), and messages formats (used by nodes of dispatcher tree)
type LogConfig struct {
	Constraints    LogLevelConstraints      // General log level rules (>min and <max, or set of allowed levels)
	Exceptions     []*LogLevelException     // Exceptions to general rules for specific files or funcs
	RootDispatcher dispatchers.DispatcherInterface // Root of output tree
	LogType        LoggerType
	LoggerData     interface{}
}

func NewConfig(
	constraints LogLevelConstraints, 
	exceptions []*LogLevelException, 
	rootDispatcher dispatchers.DispatcherInterface,
	logType LoggerType,
	logData interface{}) (*LogConfig, error) {
	if constraints == nil {
		return nil, errors.New("Constraints can not be nil")
	}
	if rootDispatcher == nil {
		return nil, errors.New("RootDispatcher can not be nil")
	}
	
	config := new(LogConfig)
	config.Constraints = constraints
	config.Exceptions = exceptions
	config.RootDispatcher = rootDispatcher
	config.LogType = logType
	config.LoggerData = logData
	
	return config, nil
}

// IsAllowed returns true if logging with specified log level is allowed in current context.
// If any of exception patterns match current context, then exception constraints are applied. Otherwise,
// the general constraints are used.
func (config *LogConfig) IsAllowed(level LogLevel, context *LogContext) bool {
	allowed := config.Constraints.IsAllowed(level) // General rule

	// Exceptions:

	for _, exception := range config.Exceptions {
		if exception.MatchesContext(context) {
			return exception.IsAllowed(level)
		}
	}

	return allowed
}