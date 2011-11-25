// Package dispatcher implements log dispatching functionality.
// It allows to filter, duplicate, customize the flow of log streams.
package dispatchers

import (
	. "sealog/common"
	"os"
	"io"
	"fmt"
)

// A DispatcherInterface is used to dispatch message to all underlying receivers.
// Dispatch logic depends on given context and log level. Any errors are reported using errorFunc.
type DispatcherInterface interface {
	Dispatch(message string, level LogLevel, context *LogContext, errorFunc func(err os.Error))
}

type dispatcher struct {
	writers     []io.Writer
	dispatchers []DispatcherInterface
}

// Creates a dispatcher which dispatches data to a list of receivers. 
// Each receiver should be either a Dispatcher or io.Writer, otherwise an error will be returned
func createDispatcher(receivers []interface{}) (*dispatcher, os.Error) {
	if receivers == nil || len(receivers) == 0 {
		return nil, os.NewError("Receivers can not be nil or empty")
	}

	disp := &dispatcher{make([]io.Writer, 0), make([]DispatcherInterface, 0)}

	for _, receiver := range receivers {
		ioWriter, ok := receiver.(io.Writer)
		if ok {
			disp.writers = append(disp.writers, ioWriter)
			continue
		}

		dispInterface, ok := receiver.(DispatcherInterface)
		if ok {
			disp.dispatchers = append(disp.dispatchers, dispInterface)
			continue
		}

		return nil, os.NewError("Method can receive either io.Writer or DispatcherInterface")
	}

	return disp, nil
}

func (disp *dispatcher) Dispatch(message string, level LogLevel, context *LogContext, errorFunc func(err os.Error)) {
	for _, writer := range disp.writers {
		_, err := writer.Write([]byte(message))
		if err != nil {
			errorFunc(err)
		}
	}

	for _, dispInterface := range disp.dispatchers {
		dispInterface.Dispatch(message, level, context, errorFunc)
	}
}

func (disp *dispatcher) Writers() []io.Writer {
	return disp.writers
}

func (disp *dispatcher) Dispatchers() []DispatcherInterface {
	return disp.dispatchers
}

func (disp *dispatcher) String() string {
	str := "    ->Dispatchers:"

	if len(disp.dispatchers) == 0 {
		str += "none\n"
	} else {
		str += "\n"

		for _, disp := range disp.dispatchers {
			str += fmt.Sprintf("        ->%s", disp)
		}
	}

	str += "    ->Writers:"

	if len(disp.writers) == 0 {
		str += "none\n"
	} else {
		str += "\n"

		for _, writer := range disp.writers {
			str += fmt.Sprintf("        ->%s", writer)
		}
	}

	return str
}
