// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"io"
	"io/ioutil"
	"fmt"
	"strconv"
	"bufio"
	"bytes"
	"errors"
)

func main() {
	fmt.Println("Hello World!")
}

func CountSequencedRowsInFile(filePath string) (int64, error) {
	bts, err := ioutil.ReadFile(filePath)
	if err != nil {
		 return 0, err
	}
	
	bufReader := bufio.NewReader(bytes.NewBuffer(bts))
	
	var gotCounter int64
	for ;; {
		line, _, bufErr := bufReader.ReadLine()
		if bufErr != nil && bufErr != io.EOF {
			return 0, bufErr
		}

		lineString := string(line)
		if lineString == "" {
			break
		}

		intVal, atoiErr := strconv.ParseInt(lineString, 10, 64)
		if atoiErr != nil {
			return 0, atoiErr
		}
		
		if intVal != gotCounter {
			return 0, errors.New(fmt.Sprintf("Wrong order: %d Expected: %d\n", intVal, gotCounter))
		}
				
		gotCounter++		
	}
	
	return gotCounter, nil
}