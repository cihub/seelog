package dispatchers

import (
	"sealog/common"
	"os"
)

// A FilterDispatcher writes the given message to underlying receivers only if message log level 
// is in the allowed list.
type FilterDispatcher struct {
	*dispatcher
	allowList map[common.LogLevel]bool
}

// NewFilterDispatcher creates a new FilterDispatcher using a list of allowed levels. 
func NewFilterDispatcher(receivers []interface{}, allowList ...common.LogLevel) (*FilterDispatcher, os.Error) {
	disp, err := createDispatcher(receivers)
	if err != nil {
		return nil, err
	}

	allows := make(map[common.LogLevel]bool)
	for _, allowLevel := range allowList {
		allows[allowLevel] = true
	}

	return &FilterDispatcher{disp, allows}, nil
}

func (this FilterDispatcher) Dispatch(message string, level common.LogLevel, context *common.LogContext, errorFunc func(err os.Error)) {
	isAllowed, ok := this.allowList[level]
	if ok && isAllowed {
		this.dispatcher.Dispatch(message, level, context, errorFunc)
	}
}
