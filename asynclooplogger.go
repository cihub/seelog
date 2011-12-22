// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sealog

import (
	cfg "github.com/cihub/sealog/config"
)

// AsyncLoopLogger represents asynchronous logger which processes the log queue in
// a 'for' loop
type AsyncLoopLogger struct {
	asyncLogger
}

// NewAsyncLoopLogger creates a new asynchronous loop logger
func NewAsyncLoopLogger(config *cfg.LogConfig) (*AsyncLoopLogger){
	
	asnLoopLogger := new(AsyncLoopLogger)
	
	asnLoopLogger.asyncLogger = *newAsyncLogger(config)
	
	go asnLoopLogger.processQueue()
	
	return asnLoopLogger
}

func (asnLoopLogger *AsyncLoopLogger) processQueue() {
	for !asnLoopLogger.closed {
		asnLoopLogger.queueHasElements.L.Lock()
		for asnLoopLogger.msgQueue.Len() == 0 && !asnLoopLogger.closed {
	   		asnLoopLogger.queueHasElements.Wait()
		}
		
		if asnLoopLogger.closed{
			break
		}

    	asnLoopLogger.processQueueElement()

		asnLoopLogger.queueHasElements.L.Unlock()
	}
}
