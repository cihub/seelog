package dispatchers

import (
	"testing"
	"os"
	. "sealog/common"
	. "sealog/test"
)

func TestFilterDispatcher_Passing(t *testing.T) {
	writer, _ := NewBytesVerfier(t)
	filter, err := NewFilterDispatcher([]interface{}{writer}, TraceLvl)
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
	writer.ExpectBytes(bytes)
	filter.Dispatch(string(bytes), TraceLvl, context, func(err os.Error) {})
	writer.MustNotExpect()
}

func TestFilterDispatcher_Denying(t *testing.T) {
	writer, _ := NewBytesVerfier(t)
	filter, err := NewFilterDispatcher([]interface{}{writer})
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
	filter.Dispatch(string(bytes), TraceLvl, context, func(err os.Error) {})
}
