package dispatchers

import (
	"os"
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
