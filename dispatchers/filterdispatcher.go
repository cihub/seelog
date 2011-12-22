// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dispatchers

import (
	"fmt"

	. "github.com/cihub/sealog/common"
	"github.com/cihub/sealog/format"
)

// A FilterDispatcher writes the given message to underlying receivers only if message log level 
// is in the allowed list.
type FilterDispatcher struct {
	*dispatcher
	allowList map[LogLevel]bool
}

// NewFilterDispatcher creates a new FilterDispatcher using a list of allowed levels. 
func NewFilterDispatcher(formatter *format.Formatter, receivers []interface{}, allowList ...LogLevel) (*FilterDispatcher, error) {
	disp, err := createDispatcher(formatter, receivers)
	if err != nil {
		return nil, err
	}

	allows := make(map[LogLevel]bool)
	for _, allowLevel := range allowList {
		allows[allowLevel] = true
	}

	return &FilterDispatcher{disp, allows}, nil
}

func (filter *FilterDispatcher) Dispatch(message string, level LogLevel, context *LogContext, errorFunc func(err error)) {
	isAllowed, ok := filter.allowList[level]
	if ok && isAllowed {
		filter.dispatcher.Dispatch(message, level, context, errorFunc)
	}
}

func (filter *FilterDispatcher) String() string {
	return fmt.Sprintf("FilterDispatcher ->\n%s", filter.dispatcher)
}
