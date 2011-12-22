// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package writers

import (
	"io"
	"testing"
)

func TestRollingFileWriter(t *testing.T) {
	newFileWriterTester(rollingFileWriterTests, rollingFileWriterGetter, t).test()
}

//===============================================================

func rollingFileWriterGetter(testCase *fileWriterTestCase) (io.Writer, error) {
	if testCase.rollingType == Size {
		return NewRollingFileWriterSize(testCase.fileName, testCase.fileSize, testCase.maxRolls)
	} else if testCase.rollingType == Date {
		return NewRollingFileWriterDate(testCase.fileName, testCase.datePattern)
	}

	panic("Incorrect rollingType")
}

//===============================================================

var rollingFileWriterTests []*fileWriterTestCase = []*fileWriterTestCase{
	createRollingSizeFileWriterTestCase([]string{}, "log.txt", 10, 10, 1, []string{"log.txt"}),
	createRollingSizeFileWriterTestCase([]string{}, "log.txt", 10, 10, 2, []string{"log.txt", "log.txt.1"}),
	createRollingSizeFileWriterTestCase([]string{"log.txt.1"}, "log.txt", 10, 10, 2, []string{"log.txt", "log.txt.1", "log.txt.2"}),
	createRollingSizeFileWriterTestCase([]string{"log.txt.1"}, "log.txt", 10, 1, 2, []string{"log.txt", "log.txt.2"}),
	createRollingSizeFileWriterTestCase([]string{}, "log.txt", 10, 1, 2, []string{"log.txt", "log.txt.1"}),
	createRollingSizeFileWriterTestCase([]string{"log.txt.9"}, "log.txt", 10, 1, 2, []string{"log.txt", "log.txt.10"}),
	createRollingSizeFileWriterTestCase([]string{"log.txt.a", "log.txt.1b"}, "log.txt", 10, 1, 2, []string{"log.txt", "log.txt.1", "log.txt.a", "log.txt.1b"}),

	createRollingSizeFileWriterTestCase([]string{}, `dir/log.txt`, 10, 10, 1, []string{`dir/log.txt`}),
	createRollingSizeFileWriterTestCase([]string{}, `dir/log.txt`, 10, 10, 2, []string{`dir/log.txt`, `dir/log.txt.1`}),
	createRollingSizeFileWriterTestCase([]string{`dir/dir/log.txt.1`}, `dir/dir/log.txt`, 10, 10, 2, []string{`dir/dir/log.txt`, `dir/dir/log.txt.1`, `dir/dir/log.txt.2`}),
	createRollingSizeFileWriterTestCase([]string{`dir/dir/dir/log.txt.1`}, `dir/dir/dir/log.txt`, 10, 1, 2, []string{`dir/dir/dir/log.txt`, `dir/dir/dir/log.txt.2`}),
	createRollingSizeFileWriterTestCase([]string{}, `./log.txt`, 10, 1, 2, []string{`log.txt`, `log.txt.1`}),
	createRollingSizeFileWriterTestCase([]string{`././././log.txt.9`}, `log.txt`, 10, 1, 2, []string{`log.txt`, `log.txt.10`}),
	createRollingSizeFileWriterTestCase([]string{"dir/dir/log.txt.a", "dir/dir/log.txt.1b"}, "dir/dir/log.txt", 10, 1, 2, []string{"dir/dir/log.txt", "dir/dir/log.txt.1", "dir/dir/log.txt.a", "dir/dir/log.txt.1b"}),

	createRollingSizeFileWriterTestCase([]string{}, `././dir/log.txt`, 10, 10, 1, []string{`dir/log.txt`}),
	createRollingSizeFileWriterTestCase([]string{}, `././dir/log.txt`, 10, 10, 2, []string{`dir/log.txt`, `dir/log.txt.1`}),
	createRollingSizeFileWriterTestCase([]string{`././dir/dir/log.txt.1`}, `dir/dir/log.txt`, 10, 10, 2, []string{`dir/dir/log.txt`, `dir/dir/log.txt.1`, `dir/dir/log.txt.2`}),
	createRollingSizeFileWriterTestCase([]string{`././dir/dir/dir/log.txt.1`}, `dir/dir/dir/log.txt`, 10, 1, 2, []string{`dir/dir/dir/log.txt`, `dir/dir/dir/log.txt.2`}),
	createRollingSizeFileWriterTestCase([]string{}, `././log.txt`, 10, 1, 2, []string{`log.txt`, `log.txt.1`}),
	createRollingSizeFileWriterTestCase([]string{`././././log.txt.9`}, `log.txt`, 10, 1, 2, []string{`log.txt`, `log.txt.10`}),
	createRollingSizeFileWriterTestCase([]string{"././dir/dir/log.txt.a", "././dir/dir/log.txt.1b"}, "dir/dir/log.txt", 10, 1, 2, []string{"dir/dir/log.txt", "dir/dir/log.txt.1", "dir/dir/log.txt.a", "dir/dir/log.txt.1b"}),

	//createRollingDateFileWriterTestCase([]string{}, "log.txt", "02.01.2006", 1, []string{}),
	//createRollingDateFileWriterTestCase([]string{}, "log.txt", "02.01.2006.000000", 2, []string{}),
}
