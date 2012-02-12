// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seelog

import (
	. "github.com/cihub/seelog/common"
	cfg "github.com/cihub/seelog/config"
)

// SyncLogger performs logging in the same goroutine where 'Trace/Debug/...'
// func was called
type SyncLogger struct {
	commonLogger 
}

// NewSyncLogger creates a new synchronous logger
func NewSyncLogger(config *cfg.LogConfig) (*SyncLogger){
	syncLogger := new(SyncLogger)
	
	syncLogger.commonLogger = *newCommonLogger(config, syncLogger)
	
	return syncLogger
}

func (cLogger *SyncLogger) innerLog(
    level LogLevel, 
	context *LogContext,
	format string, 
	params []interface{}) {
	
	cLogger.processLogMsg(level, format, params, context)
}

func (syncLogger *SyncLogger) Close() {
	syncLogger.config.RootDispatcher.Close()
}

func (syncLogger *SyncLogger) Flush() {
	syncLogger.config.RootDispatcher.Flush()
}