// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	library "github.com/cihub/sealog/examples/library/library"
	"fmt"
)

var console ConsoleWriter

type ConsoleWriter struct{}

func (cons ConsoleWriter) Print(p string) (n int, err error) {
	return cons.Write([]byte(p))
}

func (cons ConsoleWriter) Write(p []byte) (n int, err error) {
	fmt.Println(string(p))
	return len(p), nil
}

func calcF() {
	x := 1
	y := 2
	console.Print("Calculating F")
	result := library.CalculateF(x, y)
	console.Print(fmt.Sprintf("Got F = %d", result))
}

func main() {
	library.SetLogWriter(console)
	calcF()
	library.FlushLog()
}
