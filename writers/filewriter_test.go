// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package writers

import (
	"github.com/cihub/sealog/test"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

const (
	WriteMessageLen = 10
)

var bytesFileTest = []byte(strings.Repeat("A", WriteMessageLen))

func TestSimpleFileWriter(t *testing.T) {
	newFileWriterTester(simpleFileWriterTests, simpleFileWriterGetter, t).test()
}

//===============================================================

func simpleFileWriterGetter(testCase *fileWriterTestCase) (io.Writer, error) {
	return NewFileWriter(testCase.fileName)
}

//===============================================================

type fileWriterTestCase struct {
	files []string

	fileName    string
	rollingType RollingTypes
	fileSize    int64
	maxRolls    int
	datePattern string

	writeCount int

	resFiles []string
}

func createSimpleFileWriterTestCase(fileName string, writeCount int) *fileWriterTestCase {
	return &fileWriterTestCase{[]string{}, fileName, Size, 0, 0, "", writeCount, []string{fileName}}
}
func createRollingSizeFileWriterTestCase(files []string, fileName string, fileSize int64, maxRolls int, writeCount int, resFiles []string) *fileWriterTestCase {
	return &fileWriterTestCase{files, fileName, Size, fileSize, maxRolls, "", writeCount, resFiles}
}
func createRollingDateFileWriterTestCase(files []string, fileName string, datePattern string, writeCount int, resFiles []string) *fileWriterTestCase {
	return &fileWriterTestCase{files, fileName, Date, 0, 0, datePattern, writeCount, resFiles}
}

var simpleFileWriterTests []*fileWriterTestCase = []*fileWriterTestCase{
	createSimpleFileWriterTestCase("log.txt", 1),
	createSimpleFileWriterTestCase("log.txt", 50),
	createSimpleFileWriterTestCase("dir/log.txt", 1),
}

//===============================================================

type fileWriterTester struct {
	testCases   []*fileWriterTestCase
	writerGeter func(*fileWriterTestCase) (io.Writer, error)
	t           *testing.T
}

func newFileWriterTester(testCases []*fileWriterTestCase, writerGeter func(*fileWriterTestCase) (io.Writer, error), t *testing.T) *fileWriterTester {
	return &fileWriterTester{testCases, writerGeter, t}
}

func (this *fileWriterTester) test() {
	writer, err := test.NewBytesVerfier(this.t)
	if err != nil {
		this.t.Error(err)
		return
	}

	for testNum, testCase := range this.testCases {
		this.t.Logf("Start test  [%v]\n", testNum)
		fileSystemWrapperTest, err := test.NewFSTestWrapper(nil, writer, testCase.fileSize)
		if err != nil {
			this.t.Error(err)
			return
		}

		for _, filePath := range testCase.files {
			dir, _ := filepath.Split(filePath)
			err := fileSystemWrapperTest.MkdirAll(dir)
			if err != nil {
				this.t.Error(err)
				return
			}

			_, err = fileSystemWrapperTest.Create(filePath)
			if err != nil {
				this.t.Error(err)
				return
			}
		}

		fileSystemWrapper = fileSystemWrapperTest

		fileWriter, err := this.writerGeter(testCase)
		if err != nil {
			this.t.Error(err)
			return
		}

		this.performWrite(fileWriter, writer, testCase.writeCount)
		this.checkRequiredFilesExist(testCase, fileSystemWrapperTest)
		this.checkJustRequiredFilesExist(testCase, fileSystemWrapperTest)
	}
}

func (this *fileWriterTester) performWrite(fileWriter io.Writer, writer *test.BytesVerifier, count int) {
	for i := 0; i < count; i++ {
		writer.ExpectBytes(bytesFileTest)
		fileWriter.Write(bytesFileTest)
		writer.MustNotExpect()
	}
}

func (this *fileWriterTester) checkRequiredFilesExist(testCase *fileWriterTestCase, fileSystemWrapperTest *test.FileSystemTestWrapper) {
	for _, mustExistsFile := range testCase.resFiles {
		found := false
		for _, existsFile := range fileSystemWrapperTest.Files() {
			if mustExistsFile == existsFile {
				found = true
				break
			}
		}

		if !found {
			this.t.Errorf("Expected file: %v doesn't' exist", mustExistsFile)
		}
	}
}

func (this *fileWriterTester) checkJustRequiredFilesExist(testCase *fileWriterTestCase, fileSystemWrapperTest *test.FileSystemTestWrapper) {
	for _, existsFile := range fileSystemWrapperTest.Files() {
		found := false
		for _, mustExistsFile := range testCase.resFiles {
			if mustExistsFile == existsFile {
				found = true
				break
			}
		}

		if !found {
			this.t.Errorf("Unexpected file: %v", existsFile)
		}
	}
}
