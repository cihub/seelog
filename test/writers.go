// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

// BytesVerifier is a byte receiver which is used for correct input testing. 
// It allows to compare expected result and actual result in context of received bytes.
type BytesVerifier struct {
	expectedBytes   []byte // bytes that are expected to be written in next Write call
	waitingForInput bool   // true if verifier is waiting for a Write call
	writtenData     []byte // real bytes that actually were received during the last Write call
	testEnv         *testing.T
}

func NewBytesVerfier(t *testing.T) (*BytesVerifier, error) {
	if t == nil {
		return nil, errors.New("Testing environment param is nil")
	}

	verifier := new(BytesVerifier)
	verifier.testEnv = t

	return verifier, nil
}

// Write is used to check whether verifier was waiting for input and whether bytes are the same as expectedBytes.
// After Write call, waitingForInput is set to false.
func (verifier *BytesVerifier) Write(bytes []byte) (n int, err error) {
	if !verifier.waitingForInput {
		verifier.testEnv.Errorf("Unexpected input: %v", string(bytes))
		return
	}

	verifier.waitingForInput = false
	verifier.writtenData = bytes

	if verifier.expectedBytes != nil {
		if bytes == nil {
			verifier.testEnv.Errorf("Incoming 'bytes' is nil")
		} else {
			if len(bytes) != len(verifier.expectedBytes) {
				verifier.testEnv.Errorf("'Bytes' has unexpected len. Expected: %d. Got: %d. . Expected string: %q. Got: %q",
					len(verifier.expectedBytes), len(bytes), string(verifier.expectedBytes), string(bytes))
			} else {
				for i := 0; i < len(bytes); i++ {
					if verifier.expectedBytes[i] != bytes[i] {
						verifier.testEnv.Errorf("Incorrect data on position %d. Expected: %d. Got: %d. Expected string: %q. Got: %q",
							i, verifier.expectedBytes[i], bytes[i], string(verifier.expectedBytes), string(bytes))
						break
					}
				}
			}
		}
	}

	return len(bytes), nil
}

func (verifier *BytesVerifier) ExpectBytes(bytes []byte) {
	verifier.waitingForInput = true
	verifier.expectedBytes = bytes
}

func (verifier *BytesVerifier) MustNotExpect() {
	if verifier.waitingForInput {
		errorText := "Writer must not expect: "

		if verifier.expectedBytes != nil {
			errorText += "len = " + strconv.Itoa(len(verifier.expectedBytes))
			errorText += ". text = " + string(verifier.expectedBytes)
		}

		verifier.testEnv.Errorf(errorText)
	}
}

func (verifier *BytesVerifier) MustNotExpectWithDelay(delay time.Duration) {
	c := make(chan int)
	time.AfterFunc(delay, func() {
		verifier.MustNotExpect()

		c <- 1
	})

	<-c
}

func (verifier *BytesVerifier) Close() error {
	return nil
}

// NullWriter implements io.Writer inteface and does nothing, always returning a successful write result
type NullWriter struct {

}

func (this *NullWriter) Write(bytes []byte) (n int, err error) {
	return len(bytes), nil
}

func (this *NullWriter) Close() error {
	return nil
}
