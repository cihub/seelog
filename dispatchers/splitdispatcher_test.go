package dispatchers

import (
	"testing"
	"os"
	. "sealog/common"
	. "sealog/test"
)

func TestSplitDispatcher(t *testing.T) {
	writer1, _ := NewBytesVerfier(t)
	writer2, _ := NewBytesVerfier(t)
	spliter, err := NewSplitDispatcher([]interface{}{writer1, writer2})
	if err != nil {
		t.Error(err)
		return
	}

	context, err := CurrentContext()
	if err != nil {
		t.Error(err)
		return
	}

	bytes := []byte("Hello")

	writer1.ExpectBytes(bytes)
	writer2.ExpectBytes(bytes)
	spliter.Dispatch(string(bytes), TraceLvl, context, func(err os.Error) {})
	writer1.MustNotExpect()
	writer2.MustNotExpect()
}
