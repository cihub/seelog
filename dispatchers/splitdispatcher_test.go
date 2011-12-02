package dispatchers

import (
	"testing"
	"os"
	"fmt"
	. "github.com/cihub/sealog/common"
	. "github.com/cihub/sealog/test"
	"github.com/cihub/sealog/format"
)

var onlyMessageFormatForTest *format.Formatter
func init() {
	var err os.Error
	onlyMessageFormatForTest, err = format.NewFormatter("%Msg")
	if err != nil {
		fmt.Println("Can not create only message format: " + err.String())
	}
}

func TestSplitDispatcher(t *testing.T) {
	writer1, _ := NewBytesVerfier(t)
	writer2, _ := NewBytesVerfier(t)
	spliter, err := NewSplitDispatcher(onlyMessageFormatForTest, []interface{}{writer1, writer2})
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
