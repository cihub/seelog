// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dispatchers

import (
	. "github.com/cihub/sealog/common"

	"github.com/cihub/sealog/format"
	. "github.com/cihub/sealog/test"
	"testing"
)

func TestFilterDispatcher_Passing(t *testing.T) {
	writer, _ := NewBytesVerfier(t)
	filter, err := NewFilterDispatcher(onlyMessageFormatForTest, []interface{}{writer}, TraceLvl)
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
	writer.ExpectBytes(bytes)
	filter.Dispatch(string(bytes), TraceLvl, context, func(err error) {})
	writer.MustNotExpect()
}

func TestFilterDispatcher_Denying(t *testing.T) {
	writer, _ := NewBytesVerfier(t)
	filter, err := NewFilterDispatcher(format.DefaultFormatter, []interface{}{writer})
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
	filter.Dispatch(string(bytes), TraceLvl, context, func(err error) {})
}
