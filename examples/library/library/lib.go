// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testlibrary

import (
	log "github.com/cihub/sealog"
	"github.com/cihub/sealog/common"
	"errors"
	"io"
)

var logger log.LoggerInterface

func init() {
	DisableLog()
}

// calculateF is a meaningless example which just imitates some 
// heavy calculation operation and performs logging inside
func CalculateF(x, y int) int {
	logger.Info("Calculating F(%d, %d)", x, y)
	
	for i := 0; i < 10; i++ {
		logger.Trace("F calc iteration %d", i)
	}
	
	result := x + y
	
	logger.Debug("F = %d", result)
	return result
}

func DisableLog() {
	logger = log.Disabled
}

func UseLogger(newLogger log.LoggerInterface) {
	logger = newLogger
}

func SetLogWriter(writer io.Writer) error {
	if writer == nil {
		return errors.New("Nil writer")
	}
	
	newLogger, err := log.LoggerFromWriterAndLevel(writer, common.TraceLvl)
	if err != nil {
		return err
	}
	
	UseLogger(newLogger)
	return nil
}

// Call this before app shutdown
func FlushLog() {
	logger.Flush()
}