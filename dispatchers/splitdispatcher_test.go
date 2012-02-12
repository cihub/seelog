// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dispatchers

import (
	"fmt"

	. "github.com/cihub/seelog/common"
	"github.com/cihub/seelog/format"
	. "github.com/cihub/seelog/test"
	"testing"
)

var onlyMessageFormatForTest *format.Formatter

func init() {
	var err error
	onlyMessageFormatForTest, err = format.NewFormatter("%Msg")
	if err != nil {
		fmt.Println("Can not create only message format: " + err.Error())
	}
}

func TestSplitDispatcher(t *testing.T) {
	writer1, _ := NewBytesVerfier(t)
	writer2, _ := NewBytesVerfier(t)
	spliter, err := NewSplitDispatcher(onlyMessageFormatForTest, []interface{}{writer1, writer2})
	if err != nil {
		t.Error(err)
		return
	}

	context, err := CurrentContext()
	if err != nil {
		t.Error(err)
		return
	}

	bytes := []byte("Hello")

	writer1.ExpectBytes(bytes)
	writer2.ExpectBytes(bytes)
	spliter.Dispatch(string(bytes), TraceLvl, context, func(err error) {})
	writer1.MustNotExpect()
	writer2.MustNotExpect()
}
