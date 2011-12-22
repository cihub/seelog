// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sealog

import (
	cfg "github.com/cihub/sealog/config"
	"time"
)

// AsyncTimerLogger represents asynchronous logger which processes the log queue each
// 'duration' nanoseconds
type AsyncTimerLogger struct {
	asyncLogger
	interval time.Duration
}

// NewAsyncLoopLogger creates a new asynchronous loop logger
func NewAsyncTimerLogger(config *cfg.LogConfig, interval time.Duration) (*AsyncTimerLogger){
	asnTimerLogger := new(AsyncTimerLogger)
	
	asnTimerLogger.asyncLogger = *newAsyncLogger(config)
	asnTimerLogger.interval = interval
	
	go asnTimerLogger.processQueue()
	
	return asnTimerLogger
}

func (asnTimerLogger *AsyncTimerLogger) processQueue() {
	for !asnTimerLogger.closed {
		asnTimerLogger.queueHasElements.L.Lock()
		for asnTimerLogger.msgQueue.Len() == 0 && !asnTimerLogger.closed {
	   		asnTimerLogger.queueHasElements.Wait()
		}
		
		if asnTimerLogger.closed{
			break
		}
    	asnTimerLogger.processQueueElement()
	
		asnTimerLogger.queueHasElements.L.Unlock()
		
		<-time.After(asnTimerLogger.interval)
	}
}
