// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sealog

import (
	cfg "github.com/cihub/sealog/config"
	. "github.com/cihub/sealog/common"
	"fmt"
)

func buildLogString(format string, params []interface{}) string {
	var logString string
	if len(params) != 0 {
		logString = fmt.Sprintf(format, params...)
	} else {
		logString = format
	}

	return logString
}

func reportInternalError(err error) {
	fmt.Println("Sealog error: " + err.Error())
}

// LoggerInterface represents structs capable of logging Sealog messages
type LoggerInterface interface {
	Trace(format string, params ...interface{})
	Debug(format string, params ...interface{})
	Info(format string, params ...interface{})
	Warn(format string, params ...interface{})
	Error(format string, params ...interface{})
	Critical(format string, params ...interface{})
	Close()
	Flush()
	Closed() bool
}

// innerLoggerInterface is an internal logging interface
type innerLoggerInterface interface {
	log(level LogLevel, format string, params []interface{})
}


// [file path][func name][level] -> [allowed]
type allowedContextCache map[string]map[string]map[string]bool

// commonLogger contains all common data needed for logging and contains methods used to log messages.
type commonLogger struct {
	config *cfg.LogConfig // Config used for logging
	contextCache allowedContextCache // Caches whether log is enabled for specific "full path-func name-level" sets
	closed bool // 'true' when all writers are closed, all data is flushed, logger is unusable.
	unusedLevels []bool 
	innerLogger innerLoggerInterface
}

func newCommonLogger(config *cfg.LogConfig, internalLogger innerLoggerInterface) (*commonLogger) {
	cLogger := new(commonLogger)
	
	cLogger.config = config
	cLogger.contextCache = make(map[string]map[string]map[string]bool)
	cLogger.unusedLevels = make([]bool, Off)
	cLogger.fillUnusedLevels()
	cLogger.innerLogger = internalLogger
	
	return cLogger
}

func (syncLogger *commonLogger) Trace(format string, params ...interface{}) {
	if syncLogger.unusedLevels[TraceLvl] {
		return
	}
	
	syncLogger.innerLogger.log(TraceLvl, format, params)
}

func (syncLogger *commonLogger) Debug(format string, params ...interface{}) {
	if syncLogger.unusedLevels[DebugLvl] {
		return
	}
	
	syncLogger.innerLogger.log(DebugLvl, format, params)
}

func (syncLogger *commonLogger) Info(format string, params ...interface{}) {
	if syncLogger.unusedLevels[InfoLvl] {
		return
	}
	
	syncLogger.innerLogger.log(InfoLvl, format, params)
}

func (syncLogger *commonLogger) Warn(format string, params ...interface{}) {
	if syncLogger.unusedLevels[WarnLvl] {
		return
	}
	
	syncLogger.innerLogger.log(WarnLvl, format, params)
}

func (syncLogger *commonLogger) Error(format string, params ...interface{}) {
	if syncLogger.unusedLevels[ErrorLvl] {
		return
	}
	
	syncLogger.innerLogger.log(ErrorLvl, format, params)
}

func (syncLogger *commonLogger) Critical(format string, params ...interface{}) {
	if syncLogger.unusedLevels[CriticalLvl] {
		return
	}
	
	syncLogger.innerLogger.log(CriticalLvl, format, params)
}


func (cLogger *commonLogger) Closed() bool {
	return cLogger.closed
}

func (cLogger *commonLogger) fillUnusedLevels() {
	for i:= 0; i < len(cLogger.unusedLevels); i++ {
		cLogger.unusedLevels[i] = true
	}
	
	cLogger.fillUnusedLevelsByContraint(cLogger.config.Constraints)
	
	for _, exception := range cLogger.config.Exceptions {
		cLogger.fillUnusedLevelsByContraint(exception)
	}
}

func (cLogger *commonLogger) fillUnusedLevelsByContraint(constraint LogLevelConstraints) {
	for i:= 0; i < len(cLogger.unusedLevels); i++ {
		if constraint.IsAllowed(LogLevel(i)) {
			cLogger.unusedLevels[i] = false
		}
	}
}

func (cLogger *commonLogger) processLogMsg(
    level LogLevel, 
	format string, 
	params []interface{},
	context *LogContext) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	if cLogger.config.IsAllowed(level, context) {
		message := buildLogString(format, params)
		cLogger.config.RootDispatcher.Dispatch(message, level, context, reportInternalError)
	}
}


func (cLogger *commonLogger) isAllowed(level LogLevel, context *LogContext) bool {
	funcMap, ok := cLogger.contextCache[context.FullPath()]
	if !ok {
		funcMap = make(map[string]map[string]bool, 0)
		cLogger.contextCache[context.FullPath()] = funcMap
	}
	
	levelMap, ok := funcMap[context.Func()]
	if !ok {
		levelMap = make(map[string]bool, 0)
		funcMap[context.Func()] = levelMap
	}
	
	isAllowValue, ok := levelMap[level.String()]
	if !ok {
		isAllowValue = cLogger.config.IsAllowed(level, context)
		levelMap[level.String()] = isAllowValue
	}
	
	return isAllowValue
}