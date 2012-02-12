// Copyright (c) 2012 - Cloud Instruments Co. Ltd.
// 
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met: 
// 
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer. 
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution. 
// 
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

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
