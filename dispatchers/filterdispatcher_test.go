package dispatchers

import (
	"testing"
	"sealog/common"
	"os"
)

func TestFilterDispatcher_Passing(t *testing.T) {
	testEnv = t

	writer := new(testWriteCloser).Initialize()
	filter, err := NewFilterDispatcher([]interface{}{writer}, common.TraceLvl)
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
	writer.expectBytes(bytes)
	filter.Dispatch(string(bytes), common.TraceLvl, context, func(err os.Error) {})
	writer.mustNotExpect()
}

func TestFilterDispatcher_Denying(t *testing.T) {
	testEnv = t

	writer := new(testWriteCloser).Initialize()
	filter, err := NewFilterDispatcher([]interface{}{writer})
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
	filter.Dispatch(string(bytes), common.TraceLvl, context, func(err os.Error) {})
}
