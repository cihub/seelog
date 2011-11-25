package writers

import (
	"os"
)

// ConsoleWriter is used to write to console
type ConsoleWriter struct {

}

// Creates a new console writer. Returns error, if the console writer couldn't be created.
func NewConsoleWriter() (writer *ConsoleWriter, err os.Error) {
	newWriter := new(ConsoleWriter)

	return newWriter, nil
}

// Create folder and file on WriteLog/Write first call
func (console *ConsoleWriter) Write(bytes []byte) (n int, err os.Error) {
	return n, nil
}

func (console *ConsoleWriter) String() string {
	return "Console writer"
}
