// Package sealog implements logging functionality with flexible dispatching, filtering, and formatting.
package sealog

import (
	"sealog/common"
	"fmt"
	"os"
)

var currentConfig *LogConfig

func reportInternalError(err os.Error) {
	fmt.Println("Sealog error: " + err.String())
}

func log(config *LogConfig, level common.LogLevel, format string, params []interface{}) {
	defer func() {
		if err := recover(); err != nil {
			reportInternalError(os.NewError(fmt.Sprintf("%v", err)))
		}
	}()

	context, err := common.CurrentContext()
	if err != nil {
		reportInternalError(err)
		return
	}

	if config.IsAllowed(level, context) {
		dispatcher := config.RootDispatcher
		message := buildLogString(format, params)
		dispatcher.Dispatch(message, level, context, reportInternalError)
	}
}

func buildLogString(format string, params []interface{}) string {
	var logString string
	if len(params) != 0 {
		logString = fmt.Sprintf(format, params...)
	} else {
		logString = format
	}

	return logString
}

// Loads config that will be used until next SetConfig call. 
func SetConfig(config *LogConfig) {
	currentConfig = config
}

// Trace formats message according to format specifier and writes to default logger with log level = Trace
func Trace(format string, params ...interface{}) {
	log(currentConfig, common.TraceLvl, format, params)
}

// Debug formats message according to format specifier and writes to default logger with log level = Debug
func Debug(format string, params ...interface{}) {
	log(currentConfig, common.DebugLvl, format, params)
}

// Info formats message according to format specifier and writes to default logger with log level = Info
func Info(format string, params ...interface{}) {
	log(currentConfig, common.InfoLvl, format, params)
}

// Warn formats message according to format specifier and writes to default logger with log level = Warn
func Warn(format string, params ...interface{}) {
	log(currentConfig, common.WarnLvl, format, params)
}

// Error formats message according to format specifier and writes to default logger with log level = Error
func Error(format string, params ...interface{}) {
	log(currentConfig, common.ErrorLvl, format, params)
}

// Critical formats message according to format specifier and writes to default logger with log level = Critical
func Critical(format string, params ...interface{}) {
	log(currentConfig, common.CriticalLvl, format, params)
}
