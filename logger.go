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

package seelog

import (
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
	fmt.Println("Seelog error: " + err.Error())
}

// LoggerInterface represents structs capable of logging Seelog messages
type LoggerInterface interface {
	Trace(format string, params ...interface{})
	Debug(format string, params ...interface{})
	Info(format string, params ...interface{})
	Warn(format string, params ...interface{})
	Error(format string, params ...interface{})
	Critical(format string, params ...interface{})
	
	traceWithCallDepth(callDepth int, format string, params []interface{})
	debugWithCallDepth(callDepth int, format string, params []interface{})
	infoWithCallDepth(callDepth int, format string, params []interface{})
	warnWithCallDepth(callDepth int, format string, params []interface{})
	errorWithCallDepth(callDepth int, format string, params []interface{})
	criticalWithCallDepth(callDepth int, format string, params []interface{})
	
	Close()
	Flush()
	Closed() bool
}

// innerLoggerInterface is an internal logging interface
type innerLoggerInterface interface {
	innerLog(level LogLevel, context logContextInterface, format string, params []interface{})
	Flush()
}


// [file path][func name][level] -> [allowed]
type allowedContextCache map[string]map[string]map[LogLevel]bool

// commonLogger contains all common data needed for logging and contains methods used to log messages.
type commonLogger struct {
	config *logConfig // Config used for logging
	contextCache allowedContextCache // Caches whether log is enabled for specific "full path-func name-level" sets
	closed bool // 'true' when all writers are closed, all data is flushed, logger is unusable.
	unusedLevels []bool 
	innerLogger innerLoggerInterface
}

func newCommonLogger(config *logConfig, internalLogger innerLoggerInterface) (*commonLogger) {
	cLogger := new(commonLogger)
	
	cLogger.config = config
	cLogger.contextCache = make(allowedContextCache)
	cLogger.unusedLevels = make([]bool, Off)
	cLogger.fillUnusedLevels()
	cLogger.innerLogger = internalLogger
	
	return cLogger
}

func (cLogger *commonLogger) Trace(format string, params ...interface{}) {
	cLogger.traceWithCallDepth(loggerFuncCallDepth, format, params)
}

func (cLogger *commonLogger) Debug(format string, params ...interface{}) {
	cLogger.debugWithCallDepth(loggerFuncCallDepth, format, params)
}

func (cLogger *commonLogger) Info(format string, params ...interface{}) {
	cLogger.infoWithCallDepth(loggerFuncCallDepth, format, params)
}

func (cLogger *commonLogger) Warn(format string, params ...interface{}) {
	cLogger.warnWithCallDepth(loggerFuncCallDepth, format, params)
}

func (cLogger *commonLogger) Error(format string, params ...interface{}) {
	cLogger.errorWithCallDepth(loggerFuncCallDepth, format, params)
}

func (cLogger *commonLogger) Critical(format string, params ...interface{}) {
	cLogger.criticalWithCallDepth(loggerFuncCallDepth, format, params)
}

func (cLogger *commonLogger) traceWithCallDepth(callDepth int, format string, params []interface{}) {
	cLogger.log(TraceLvl, format, params, callDepth)
}

func (cLogger *commonLogger) debugWithCallDepth(callDepth int, format string, params []interface{}) {
	cLogger.log(DebugLvl, format, params, callDepth)
}

func (cLogger *commonLogger) infoWithCallDepth(callDepth int, format string, params []interface{}) {
	cLogger.log(InfoLvl, format, params, callDepth)
}

func (cLogger *commonLogger) warnWithCallDepth(callDepth int, format string, params []interface{}) {
	cLogger.log(WarnLvl, format, params, callDepth)
}

func (cLogger *commonLogger) errorWithCallDepth(callDepth int, format string, params []interface{}) {
	cLogger.log(ErrorLvl, format, params, callDepth)
}

func (cLogger *commonLogger) criticalWithCallDepth(callDepth int, format string, params []interface{}) {
	cLogger.log(CriticalLvl, format, params, callDepth)
	cLogger.innerLogger.Flush()
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

func (cLogger *commonLogger) fillUnusedLevelsByContraint(constraint logLevelConstraints) {
	for i:= 0; i < len(cLogger.unusedLevels); i++ {
		if constraint.IsAllowed(LogLevel(i)) {
			cLogger.unusedLevels[i] = false
		}
	}
}

// stackCallDepth is used to indicate the call depth of 'log' func.
// This depth level is used in the runtime.Caller(...) call. See 
// common_context.go -> specificContext, extractCallerInfo for details.
func (cLogger *commonLogger) log(
    level LogLevel, 
	format string, 
	params []interface{},
	stackCallDepth int) {
	
	if cLogger.Closed() {
		return
	}
	
	if cLogger.unusedLevels[level] {
		return
	}
	
	context, _ := specificContext(stackCallDepth)
	
	// Context errors are not reported because there are situations
	// in which context errors are normal Seelog usage cases. For 
	// example in executables with stripped symbols.
	// Error contexts are returned instead. See common_context.go.
	/*if err != nil {
		reportInternalError(err)
		return
	}*/
	
	cLogger.innerLogger.innerLog(level, context, format, params)
}


func (cLogger *commonLogger) processLogMsg(
    level LogLevel, 
	format string, 
	params []interface{},
	context logContextInterface) {

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


func (cLogger *commonLogger) isAllowed(level LogLevel, context logContextInterface) bool {
	funcMap, ok := cLogger.contextCache[context.FullPath()]
	if !ok {
		funcMap = make(map[string]map[LogLevel]bool, 0)
		cLogger.contextCache[context.FullPath()] = funcMap
	}
	
	levelMap, ok := funcMap[context.Func()]
	if !ok {
		levelMap = make(map[LogLevel]bool, 0)
		funcMap[context.Func()] = levelMap
	}
	
	isAllowValue, ok := levelMap[level]
	if !ok {
		isAllowValue = cLogger.config.IsAllowed(level, context)
		levelMap[level] = isAllowValue
	}
	
	return isAllowValue
}
