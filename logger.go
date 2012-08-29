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

func reportInternalError(err error) {
	fmt.Println("Seelog error: " + err.Error())
}

// LoggerInterface represents structs capable of logging Seelog messages
type LoggerInterface interface {
	Tracef(format string, params ...interface{})
	Debugf(format string, params ...interface{})
	Infof(format string, params ...interface{})
	Warnf(format string, params ...interface{})
	Errorf(format string, params ...interface{})
	Criticalf(format string, params ...interface{})
	
	Trace(v ...interface{})
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	
	traceWithCallDepth(callDepth int, message *logMessage)
	debugWithCallDepth(callDepth int, message *logMessage)
	infoWithCallDepth(callDepth int, message *logMessage)
	warnWithCallDepth(callDepth int, message *logMessage)
	errorWithCallDepth(callDepth int, message *logMessage)
	criticalWithCallDepth(callDepth int, message *logMessage)
	
	Close()
	Flush()
	Closed() bool
}

// innerLoggerInterface is an internal logging interface
type innerLoggerInterface interface {
	innerLog(level LogLevel, context logContextInterface, message *logMessage)
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

func (cLogger *commonLogger) Tracef(format string, params ...interface{}) {
	cLogger.traceWithCallDepth(loggerFuncCallDepth, newFormattedLogMessage(format, params))
}

func (cLogger *commonLogger) Debugf(format string, params ...interface{}) {
	cLogger.debugWithCallDepth(loggerFuncCallDepth, newFormattedLogMessage(format, params))
}

func (cLogger *commonLogger) Infof(format string, params ...interface{}) {
	cLogger.infoWithCallDepth(loggerFuncCallDepth, newFormattedLogMessage(format, params))
}

func (cLogger *commonLogger) Warnf(format string, params ...interface{}) {
	cLogger.warnWithCallDepth(loggerFuncCallDepth, newFormattedLogMessage(format, params))
}

func (cLogger *commonLogger) Errorf(format string, params ...interface{}) {
	cLogger.errorWithCallDepth(loggerFuncCallDepth, newFormattedLogMessage(format, params))
}

func (cLogger *commonLogger) Criticalf(format string, params ...interface{}) {
	cLogger.criticalWithCallDepth(loggerFuncCallDepth, newFormattedLogMessage(format, params))
}

func (cLogger *commonLogger) Trace(v ...interface{}) {
	cLogger.traceWithCallDepth(loggerFuncCallDepth, newLogMessage(v))
}

func (cLogger *commonLogger) Debug(v ...interface{}) {
	cLogger.debugWithCallDepth(loggerFuncCallDepth, newLogMessage(v))
}

func (cLogger *commonLogger) Info(v ...interface{}) {
	cLogger.infoWithCallDepth(loggerFuncCallDepth, newLogMessage(v))
}

func (cLogger *commonLogger) Warn(v ...interface{}) {
	cLogger.warnWithCallDepth(loggerFuncCallDepth, newLogMessage(v))
}

func (cLogger *commonLogger) Error(v ...interface{}) {
	cLogger.errorWithCallDepth(loggerFuncCallDepth, newLogMessage(v))
}

func (cLogger *commonLogger) Critical(v ...interface{}) {
	cLogger.criticalWithCallDepth(loggerFuncCallDepth, newLogMessage(v))
}

func (cLogger *commonLogger) traceWithCallDepth(callDepth int, message *logMessage) {
	cLogger.log(TraceLvl, message, callDepth)
}

func (cLogger *commonLogger) debugWithCallDepth(callDepth int, message *logMessage) {
	cLogger.log(DebugLvl, message, callDepth)
}

func (cLogger *commonLogger) infoWithCallDepth(callDepth int, message *logMessage) {
	cLogger.log(InfoLvl, message, callDepth)
}

func (cLogger *commonLogger) warnWithCallDepth(callDepth int, message *logMessage) {
	cLogger.log(WarnLvl, message, callDepth)
}

func (cLogger *commonLogger) errorWithCallDepth(callDepth int, message *logMessage) {
	cLogger.log(ErrorLvl, message, callDepth)
}

func (cLogger *commonLogger) criticalWithCallDepth(callDepth int, message *logMessage) {
	cLogger.log(CriticalLvl, message, callDepth)
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
	message *logMessage,
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
	
	cLogger.innerLogger.innerLog(level, context, message)
}


func (cLogger *commonLogger) processLogMsg(
    level LogLevel, 
	message *logMessage,
	context logContextInterface) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	if cLogger.config.IsAllowed(level, context) {
		cLogger.config.RootDispatcher.Dispatch(message.String(), level, context, reportInternalError)
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



type logMessage struct {
	isFormatted bool
	format string
	params []interface{}
}

func newLogMessage(params []interface{}) *logMessage {
	message := new(logMessage)
	
	message.params = params
	
	return message
}

func newFormattedLogMessage(format string, params []interface{}) *logMessage {
	message := new(logMessage)
	
	message.params = params
	message.format = format
	message.isFormatted = true
	
	return message
}

func (message *logMessage) String() string {
	if message.isFormatted {
		return fmt.Sprintf(message.format, message.params...)
	} else {
		return fmt.Sprint(message.params...)
	}
	
	panic("impossible")
}