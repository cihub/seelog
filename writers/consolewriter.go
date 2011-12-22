// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package writers

import "fmt"

// ConsoleWriter is used to write to console
type ConsoleWriter struct {

}

// Creates a new console writer. Returns error, if the console writer couldn't be created.
func NewConsoleWriter() (writer *ConsoleWriter, err error) {
	newWriter := new(ConsoleWriter)

	return newWriter, nil
}

// Create folder and file on WriteLog/Write first call
func (console *ConsoleWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(string(bytes))
}

func (console *ConsoleWriter) String() string {
	return "Console writer"
}
