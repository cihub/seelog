// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sealog

import (
	"github.com/cihub/sealog/test"
	"testing"
	"strconv"
	"os"
)

func Test_Asynctimer(t *testing.T) {
	fileName := "log.log"
	count := 100
	
	os.Remove(fileName)
	
	testConfig := `
<sealog type="asynctimer" asyncinterval="100">
	<outputs formatid="msg">
		<file path="` + fileName + `"/>
	</outputs>
	<formats>
		<format id="msg" format="%Msg%n"/>
	</formats>
</sealog>`

	logger, _ := LoggerFromConfigAsBytes([]byte(testConfig))
	err := ReplaceLogger(logger)
	if err != nil {
		t.Error(err)
		return
	}
	
	for i := 0; i < count; i++ {
		Trace(strconv.Itoa(i))
	}
	
	Flush()
	
	gotCount, err := test.CountSequencedRowsInFile(fileName)
	if err != nil {
		t.Error(err)
		return
	}
	
	if int64(count) != gotCount {
		t.Errorf("Wrong count of log messages. Expected: %v, got: %v.", count, gotCount)
		return
	}
}