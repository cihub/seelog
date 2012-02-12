// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seelog

import (
	"github.com/cihub/seelog/test"
	"testing"
	"strconv"
	"os"
)

func Test_Sync(t *testing.T) {
	fileName := "log.log"
	count := 100
	
	os.Remove(fileName)
	
	testConfig := `
<seelog type="sync">
	<outputs formatid="msg">
		<file path="` + fileName + `"/>
	</outputs>
	<formats>
		<format id="msg" format="%Msg%n"/>
	</formats>
</seelog>`

	logger, _ := LoggerFromConfigAsBytes([]byte(testConfig))
	err := ReplaceLogger(logger)
	if err != nil {
		t.Error(err)
		return
	}
	defer Flush()
	
	for i := 0; i < count; i++ {
		Trace(strconv.Itoa(i))
	}
	
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