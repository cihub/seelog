// Copyright (c) 2012 - Cloud Instruments Co., Ltd.
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

package seelog

import (
	"io"
	"os"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

const (
	WriteMessageLen = 10
)

var bytesFileTest = []byte(strings.Repeat("A", WriteMessageLen))

func TestSimpleFileWriter(t *testing.T) {
	t.Logf("Starting file writer tests")
	newFileWriterTester(simplefileWriterTests, simplefileWriterGetter, t).test()
}

//===============================================================

func simplefileWriterGetter(testCase *fileWriterTestCase) (io.Writer, error) {
	return newFileWriter(testCase.fileName)
}

//===============================================================

type fileWriterTestCase struct {
	files []string

	fileName    string
	rollingType rollingTypes
	fileSize    int64
	maxRolls    int
	datePattern string

	writeCount int

	resFiles []string
}

func createSimplefileWriterTestCase(fileName string, writeCount int) *fileWriterTestCase {
	return &fileWriterTestCase{[]string{}, fileName, Size, 0, 0, "", writeCount, []string{fileName}}
}

var simplefileWriterTests []*fileWriterTestCase = []*fileWriterTestCase{
	createSimplefileWriterTestCase("log.testlog", 1),
	createSimplefileWriterTestCase("log.testlog", 50),
	createSimplefileWriterTestCase(filepath.Join("dir", "log.testlog"), 50),
}

//===============================================================

type fileWriterTester struct {
	testCases   []*fileWriterTestCase
	writerGetter func(*fileWriterTestCase) (io.Writer, error)
	t           *testing.T
}

func newFileWriterTester(testCases []*fileWriterTestCase, writerGetter func(*fileWriterTestCase) (io.Writer, error), t *testing.T) *fileWriterTester {
	return &fileWriterTester{testCases, writerGetter, t}
}

func isWriterTestFile(f os.FileInfo) bool {
	return strings.Contains(f.Name(), ".testlog")
}

func cleanupWriterTest(t *testing.T) {
	toDel, err := getDirFileNames(".", false, isWriterTestFile)

	if nil != err {
		t.Fatal("Cannot list files in test directory!")
	}

	for _, p := range toDel {
		err = os.Remove(p)

		if nil != err {
			t.Errorf("Cannot remove file %s in test directory: %s", p, err.Error())
		}
	}

	err = os.RemoveAll("dir")

	if nil != err {
		t.Errorf("Cannot remove temp test directory: %s", err.Error())
	}
}

func getWriterTestResultFiles() ([]string, error) {
	p := make([]string, 0)

	visit := func (path string, f os.FileInfo, err error) error {

		if !f.IsDir() && isWriterTestFile(f) {
  			abs, err := filepath.Abs(path)

  			if err != nil {
				return fmt.Errorf("filepath.Abs failed for %s", path)
			}

  			p = append(p, abs)
  		}

  		return nil
	} 

	err := filepath.Walk(".", visit)

	if nil != err {
		return nil, err
	}

	return p, nil
}

func (this *fileWriterTester) testCase(testCase *fileWriterTestCase, testNum int) {
	defer cleanupWriterTest(this.t)

	this.t.Logf("Start test  [%v]\n", testNum)

	for _, filePath := range testCase.files {
		dir, _ := filepath.Split(filePath)

		var err error

		if 0 != len(dir) {
			err = os.MkdirAll(dir, defaultDirectoryPermissions)
			if err != nil {
				this.t.Error(err)
				return
			}
		}
		_, err = os.Create(filePath)
		if err != nil {
			this.t.Error(err)
			return
		}
	}

	fileWriter, err := this.writerGetter(testCase)

	if err != nil {
		this.t.Error(err)
		return
	}

	this.performWrite(fileWriter, testCase.writeCount)

	files, err := getWriterTestResultFiles()

	if err != nil {
		this.t.Error(err)
		return
	}

	this.checkRequiredFilesExist(testCase, files)
	this.checkJustRequiredFilesExist(testCase, files)
}

func (this *fileWriterTester) test() {
	for i, tc := range this.testCases {
		cleanupWriterTest(this.t)
		this.testCase(tc, i)
	}
}

func (this *fileWriterTester) performWrite(fileWriter io.Writer, count int) {
	for i := 0; i < count; i++ {
		fileWriter.Write(bytesFileTest)
	}
}

func (this *fileWriterTester) checkRequiredFilesExist(testCase *fileWriterTestCase, files []string) {
	for _, expected := range testCase.resFiles {

		found := false
		exAbs, err := filepath.Abs(expected)

		if err != nil {
			this.t.Errorf("filepath.Abs failed for %s", expected)
		} else {
			for _, f := range files {

				if exAbs == f {
					found = true
					break
				}
			}

			if !found {
				this.t.Errorf("Expected file: %v doesn't exist\n", expected)
			}
		}

		
	}
}

func (this *fileWriterTester) checkJustRequiredFilesExist(testCase *fileWriterTestCase, files []string) {
	for _, f := range files {
		found := false
		for _, expected := range testCase.resFiles {

			exAbs, err := filepath.Abs(expected)
			if err != nil {
				this.t.Errorf("filepath.Abs failed for %s", expected)
			} else {
				if exAbs == f {
					found = true
					break
				}
			}
		}

		if !found {
			this.t.Errorf("Unexpected file: %v", f)
		}
	}
}
