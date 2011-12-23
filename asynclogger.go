// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sealog

import (
	cfg "github.com/cihub/sealog/config"
	. "github.com/cihub/sealog/common"
	"container/list"
	"sync"
	"fmt"
	"errors"
)

const (
	MaxQueueSize = 10000
)

type msgQueueItem struct {
	level LogLevel
	context *LogContext
	format string
	params []interface{}
}

// asyncLogger represents common data for all asynchronous loggers
type asyncLogger struct {
	commonLogger 
	msgQueue *list.List
	queueMutex *sync.Mutex
	queueHasElements *sync.Cond
}

// newAsyncLogger creates a new asynchronous logger
func newAsyncLogger(config *cfg.LogConfig) (*asyncLogger){
	asnLogger := new(asyncLogger)
	
	asnLogger.msgQueue = list.New()
	asnLogger.queueMutex = new(sync.Mutex)
	asnLogger.queueHasElements = sync.NewCond(new(sync.Mutex))
	
	asnLogger.commonLogger = *newCommonLogger(config, asnLogger)
	
	return asnLogger
}

func (asnLogger *asyncLogger) innerLog(
    level LogLevel, 
	context *LogContext,
	format string, 
	params []interface{}) {
		
	asnLogger.addMsgToQueue(level, context, format, params)
}

func (asnLogger *asyncLogger) Close() {
	asnLogger.queueMutex.Lock()
	if !asnLogger.closed {
		asnLogger.flushQueue()
		asnLogger.config.RootDispatcher.Flush()
		asnLogger.config.RootDispatcher.Close()
		asnLogger.queueHasElements.Broadcast()
	}
	asnLogger.queueMutex.Unlock()
}

func (asnLogger *asyncLogger) Flush() {
	asnLogger.queueMutex.Lock()
	
	if !asnLogger.closed {
		asnLogger.flushQueue()
		asnLogger.config.RootDispatcher.Flush()
	}
	asnLogger.queueMutex.Unlock()
}

func (asnLogger *asyncLogger) flushQueue() {
	asnLogger.queueHasElements.L.Lock()
	for asnLogger.msgQueue.Len() > 0 {
   		asnLogger.processQueueElement()
	}

	asnLogger.queueHasElements.L.Unlock()
}

func (asnLogger *asyncLogger) processQueueElement() {
	if asnLogger.msgQueue.Len() > 0 {
		backElement := asnLogger.msgQueue.Front()
		msg, _ := backElement.Value.(msgQueueItem)
		asnLogger.processLogMsg(msg.level, msg.format, msg.params, msg.context)
		asnLogger.msgQueue.Remove(backElement)
	}
}

func (asnLogger *asyncLogger) addMsgToQueue(level LogLevel, context *LogContext, format string, params []interface{}) {
	asnLogger.queueMutex.Lock()
	if !asnLogger.closed {
		if asnLogger.msgQueue.Len() >= MaxQueueSize {
			fmt.Printf("Sealog queue overflow: more than %v messages in the queue. Flushing.\n", MaxQueueSize)
			asnLogger.flushQueue()
		}
		
		param := params
		queueItem := msgQueueItem{level, context, format, param}
		asnLogger.msgQueue.PushBack(queueItem)
		asnLogger.queueHasElements.Broadcast()
	} else {
		err := errors.New(fmt.Sprintf("Queue closed! Cannot process element: %d %s %v", level, format, params))
		reportInternalError(err)
	}
	asnLogger.queueMutex.Unlock()
}

