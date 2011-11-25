package writers

import (
	"testing"
	"io"
	"os"
)

func TestRollingFileWriter(t *testing.T) {
	newFileWriterTester(rollingFileWriterTests, rollingFileWriterGetter, t).test()
}

//===============================================================

func rollingFileWriterGetter(testCase *fileWriterTestCase) (io.Writer, os.Error) {
	if testCase.rollingType == Size {
		return NewRollingFileWriterSize(testCase.fileName, testCase.fileSize, testCase.maxRolls), nil
	} else if testCase.rollingType == Date {
		return NewRollingFileWriterDate(testCase.fileName, testCase.datePattern), nil
	}

	panic("Incorrect rollingType")
}

//===============================================================

var rollingFileWriterTests []*fileWriterTestCase = []*fileWriterTestCase{
	createRollingSizeFileWriterTestCase([]string{}, "log.txt", 10, 10, 1, []string{"log.txt"}),
	createRollingSizeFileWriterTestCase([]string{}, "log.txt", 10, 10, 2, []string{"log.txt", "log.txt.1"}),
	createRollingSizeFileWriterTestCase([]string{"log.txt.1"}, "log.txt", 10, 10, 2, []string{"log.txt", "log.txt.1", "log.txt.2"}),
	createRollingSizeFileWriterTestCase([]string{"log.txt.1"}, "log.txt", 10, 1, 2, []string{"log.txt", "log.txt.2"}),
	createRollingSizeFileWriterTestCase([]string{}, "log.txt", 10, 0, 2, []string{"log.txt", "log.txt.1"}),
	createRollingSizeFileWriterTestCase([]string{"log.txt.9"}, "log.txt", 10, 1, 2, []string{"log.txt", "log.txt.10"}),

	//createRollingDateFileWriterTestCase([]string{}, "log.txt", "02.01.2006", 1, []string{}),
	//createRollingDateFileWriterTestCase([]string{}, "log.txt", "02.01.2006.000000", 2, []string{}),
}
