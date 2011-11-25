package dispatchers

import (
	"os"
	"fmt"
)

// A SplitDispatcher just writes the given message to underlying receivers. (Splits the message stream.)
type SplitDispatcher struct {
	*dispatcher
}

func NewSplitDispatcher(receivers []interface{}) (*SplitDispatcher, os.Error) {
	disp, err := createDispatcher(receivers)
	if err != nil {
		return nil, err
	}

	return &SplitDispatcher{disp}, nil
}

func (splitter *SplitDispatcher) String() string {
	return fmt.Sprintf("SplitDispatcher ->\n%s", splitter.dispatcher.String())
}
