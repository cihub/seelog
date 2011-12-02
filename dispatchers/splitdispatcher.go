package dispatchers

import (
	"os"
	"fmt"
	"github.com/cihub/sealog/format"
)

// A SplitDispatcher just writes the given message to underlying receivers. (Splits the message stream.)
type SplitDispatcher struct {
	*dispatcher
}

func NewSplitDispatcher(formatter *format.Formatter, receivers []interface{}) (*SplitDispatcher, os.Error) {
	disp, err := createDispatcher(formatter, receivers)
	if err != nil {
		return nil, err
	}

	return &SplitDispatcher{disp}, nil
}

func (splitter *SplitDispatcher) String() string {
	return fmt.Sprintf("SplitDispatcher ->\n%s", splitter.dispatcher.String())
}
