package dispatchers

import (
	"os"
	"testing"
	"time"
	"strconv"
	"sealog/common"
)

const (
	TEST_WRITER_NOT_EXPECT_LOGLEVEL = 255
)

var testEnv *testing.T

type testWriteCloser struct {
	expectedLog   common.LogLevel
	expectedBytes []byte
	expecting     bool

	writedData []byte

	writed chan int
}

func (this *testWriteCloser) Initialize() *testWriteCloser {
	this.writed = make(chan int, 1024)

	return this
}

func (this *testWriteCloser) Write(bytes []byte) (n int, err os.Error) {
	return this.WriteLog(bytes, TEST_WRITER_NOT_EXPECT_LOGLEVEL)
}
func (this *testWriteCloser) WriteLog(bytes []byte, level common.LogLevel) (n int, err os.Error) {
	if !this.expecting {
		testEnv.Errorf("Unexpected writing: %v", string(bytes))
		return
	}

	this.expecting = false
	this.writedData = bytes

	this.writed <- 1

	if this.expectedLog != TEST_WRITER_NOT_EXPECT_LOGLEVEL && this.expectedLog != level {
		testEnv.Errorf("Incorrect logLevel. Expected %v in %v", common.LogLevelToString(this.expectedLog), common.LogLevelToString(level))
	}

	if this.expectedBytes != nil {
		if bytes == nil {
			testEnv.Errorf("Incoming 'bytes' is nil")
		} else {
			if len(bytes) != len(this.expectedBytes) {
				testEnv.Errorf("'Bytes' has unexpected len: expected - %v, get - %v", len(this.expectedBytes), len(bytes))
			} else {
				for i := 0; i < len(bytes); i++ {
					if this.expectedBytes[i] != bytes[i] {
						testEnv.Errorf("Writed incorrect data on %v: exp %v get %v. %v %v",
							i, this.expectedBytes[i], bytes[i], this.expectedBytes, bytes)
					}
				}
			}
		}
	}

	return len(bytes), nil
}

func (this *testWriteCloser) expect(logLevel common.LogLevel, bytes []byte) {
	this.expecting = true
	this.expectedLog = logLevel
	this.expectedBytes = bytes
}

func (this *testWriteCloser) expectLog(logLevel common.LogLevel) {
	this.expecting = true
	this.expectedLog = logLevel
	this.expectedBytes = nil
}

func (this *testWriteCloser) expectBytes(bytes []byte) {
	this.expecting = true
	this.expectedBytes = bytes
	this.expectedLog = TEST_WRITER_NOT_EXPECT_LOGLEVEL
}

func (this *testWriteCloser) mustNotExpect() {
	if this.expecting {
		errorText := "Writer must not expect: "
		if this.expectedLog != TEST_WRITER_NOT_EXPECT_LOGLEVEL {
			errorText += "log = " + common.LogLevelToString(this.expectedLog)
		}

		if this.expectedBytes != nil {
			errorText += "len = " + strconv.Itoa(len(this.expectedBytes))
		}

		testEnv.Errorf(errorText)
	}
}

func (this *testWriteCloser) mustNotExpectWithDelay(delay int64) {
	c := make(chan int)
	time.AfterFunc(delay, func() {
		this.mustNotExpect()

		c <- 1
	})

	<-c
}

func (this *testWriteCloser) Close() os.Error {
	return nil
}

//=====================================================================================

type emptyWriteCloser struct {

}

func (this *emptyWriteCloser) Initialize() *emptyWriteCloser {
	return this
}

func (this *emptyWriteCloser) Write(bytes []byte) (n int, err os.Error) {
	return len(bytes), nil
}

func (this *emptyWriteCloser) Close() os.Error {
	return nil
}
