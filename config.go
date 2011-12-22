// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sealog

import (
	"bytes"
	. "github.com/cihub/sealog/common"
	"github.com/cihub/sealog/config"
	"github.com/cihub/sealog/dispatchers"
	"github.com/cihub/sealog/format"
	"io"
	"os"
)

// ConfigFromFile creates a config from file. File should contain valid sealog xml.
func ConfigFromFile(fileName string) (*config.LogConfig, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return config.ConfigFromReader(file)
}

// ConfigFromBytes creates a config from bytes stream. Bytes should contain valid sealog xml.
func ConfigFromBytes(data []byte) (*config.LogConfig, error) {
	return config.ConfigFromReader(bytes.NewBuffer(data))
}

// ConfigFromWriter creates a simple config for usage with non-Sealog systems. 
// Configures system to write to output with minimal level = minLevel.
func ConfigFromWriterAndLevel(output io.Writer, minLevel LogLevel) (*config.LogConfig, error) {
	constraints, err := NewMinMaxConstraints(minLevel, CriticalLvl)
	if err != nil {
		return nil, err
	}

	dispatcher, err := dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{output})
	if err != nil {
		return nil, err
	}

	return config.NewConfig(constraints, make([]*LogLevelException, 0), dispatcher, config.SyncLoggerType, nil)
}
