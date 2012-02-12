// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dispatchers

import (
	"testing"
	. "github.com/cihub/seelog/common"
	. "github.com/cihub/seelog/test"
	"github.com/cihub/seelog/format"
)

func TestFormattedWriter(t *testing.T) {
	formatStr := "%Level %LEVEL %Msg"
	message := "message"
	var logLevel LogLevel =  TraceLvl
	
	bytesVerifier, err := NewBytesVerfier(t)
	if err != nil {
		t.Error(err)
		return
	}
	
	formatter, err := format.NewFormatter(formatStr)
	if err != nil {
		t.Error(err)
		return
	}
	
	writer, err := NewFormattedWriter(bytesVerifier, formatter)
	if err != nil {
		t.Error(err)
		return
	}
	
	context, err := CurrentContext()
	if err != nil {
		t.Error(err)
		return
	}
	
	logMessage := formatter.Format(message, logLevel, context)
	
	bytesVerifier.ExpectBytes([]byte(logMessage))
	writer.Write(message, logLevel, context)
	bytesVerifier.MustNotExpect()
}