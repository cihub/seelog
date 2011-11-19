package dispatchers

import (
	"testing"
	"sealog/common"
	"os"
)

func TestSplitDispatcher(t *testing.T) {
	testEnv = t

	writer1 := new(testWriteCloser).Initialize()
	writer2 := new(testWriteCloser).Initialize()
	spliter, err := NewSplitDispatcher([]interface{}{writer1, writer2})
	if err != nil {
		testEnv.Error(err)
		return
	}

	context, err := common.CurrentContext()
	if err != nil {
		testEnv.Error(err)
		return
	}

	bytes := []byte("Hello")

	writer1.expectBytes(bytes)
	writer2.expectBytes(bytes)
	spliter.Dispatch(string(bytes), common.TraceLvl, context, func(err os.Error) {})
	writer1.mustNotExpect()
	writer2.mustNotExpect()
}
