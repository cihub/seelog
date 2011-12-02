package writers

import (
	"testing"
	"io"
	"github.com/cihub/sealog/test"
	"os"
)

func TestSimpleFileWriter(t *testing.T) {
	newFileWriterTester(simpleFileWriterTests, simpleFileWriterGetter, t).test()
}

//===============================================================

func simpleFileWriterGetter(testCase *fileWriterTestCase) (io.Writer, os.Error) {
	return NewFileWriter(testCase.fileName)
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
}

//===============================================================

type fileWriterTester struct {
	testCases   []*fileWriterTestCase
	writerGeter func(*fileWriterTestCase) (io.Writer, os.Error)
	t           *testing.T
}

func newFileWriterTester(testCases []*fileWriterTestCase, writerGeter func(*fileWriterTestCase) (io.Writer, os.Error), t *testing.T) *fileWriterTester {
	return &fileWriterTester{testCases, writerGeter, t}
}

func (this *fileWriterTester) test() {
	writer, err := test.NewBytesVerfier(this.t)
	if err != nil {
		this.t.Error(err)
		return
	}

	for _, testCase := range this.testCases {
		files := make([]*test.FileWrapper, 0)
		for _, fileName := range testCase.files {
			files = append(files, test.NewFileWrapper(fileName))
		}
		dir, err := test.NewDirectoryWrapper("", make([]*test.DirectoryWrapper, 0), files)
		if err != nil {
			this.t.Error(err)
			return
		}
		fileSystemWrapperTest, err := test.NewFSTestWrapper(dir, writer, testCase.fileSize)
		if err != nil {
			this.t.Error(err)
			return
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
	bytes := []byte("Hello")

	for i := 0; i < count; i++ {
		writer.ExpectBytes(bytes)
		fileWriter.Write(bytes)
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
			this.t.Errorf("Expected file: %v dosent exist", mustExistsFile)
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
