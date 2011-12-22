// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dispatchers

import (
	"fmt"
	"github.com/cihub/sealog/format"
)

// A SplitDispatcher just writes the given message to underlying receivers. (Splits the message stream.)
type SplitDispatcher struct {
	*dispatcher
}

func NewSplitDispatcher(formatter *format.Formatter, receivers []interface{}) (*SplitDispatcher, error) {
	disp, err := createDispatcher(formatter, receivers)
	if err != nil {
		return nil, err
	}

	return &SplitDispatcher{disp}, nil
}

func (splitter *SplitDispatcher) String() string {
	return fmt.Sprintf("SplitDispatcher ->\n%s", splitter.dispatcher.String())
}
