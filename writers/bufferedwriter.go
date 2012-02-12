// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package writers

import (
	"bufio"
	"errors"

	"fmt"
	"io"
	"sync"
	"time"
)

// BufferedWriter stores data in memory and flushes it every flushPeriod or when buffer is full
type BufferedWriter struct {
	flushPeriod time.Duration // data flushes interval (in microseconds)
	bufferMutex *sync.Mutex   // mutex for buffer operations syncronization
	innerWriter io.Writer     // inner writer
	buffer      *bufio.Writer // buffered wrapper for inner writer
	bufferSize  int           // max size of data chunk in bytes
}

// NewBufferedWriter creates a new buffered writer struct.
// bufferSize -- size of memory buffer in bytes
// flushPeriod -- period in which data flushes from memory buffer in milliseconds. 0 - turn off this functionality
func NewBufferedWriter(innerWriter io.Writer, bufferSize int, flushPeriod time.Duration) (*BufferedWriter, error) {

	if innerWriter == nil {
		return nil, errors.New("Argument is nil: innerWriter")
	}
	if flushPeriod < 0 {
		return nil, errors.New(fmt.Sprintf("flushPeriod can not be less than 0. Got: %d", flushPeriod))
	}

	if bufferSize <= 0 {
		return nil, errors.New(fmt.Sprintf("bufferSize can not be less or equal to 0. Got: %d", bufferSize))
	}

	buffer := bufio.NewWriterSize(innerWriter, bufferSize)

	/*if err != nil {
		return nil, err
	}*/

	newWriter := new(BufferedWriter)

	newWriter.innerWriter = innerWriter
	newWriter.buffer = buffer
	newWriter.bufferSize = bufferSize
	newWriter.flushPeriod = flushPeriod * 1e6
	newWriter.bufferMutex = new(sync.Mutex)

	if flushPeriod != 0 {
		go newWriter.flushPeriodically()
	}

	return newWriter, nil
}

func (bufWriter *BufferedWriter) writeBigChunk(bytes []byte) (n int, err error) {
	bufferedLen := bufWriter.buffer.Buffered()

	n, err = bufWriter.flushInner()
	if err != nil {
		return
	}

	written, writeErr := bufWriter.innerWriter.Write(bytes)
	return bufferedLen + written, writeErr
}

// Sends data to buffer manager. Waits until all buffers are full.
func (bufWriter *BufferedWriter) Write(bytes []byte) (n int, err error) {

	bufWriter.bufferMutex.Lock()
	defer bufWriter.bufferMutex.Unlock()

	bytesLen := len(bytes)

	if bytesLen > bufWriter.bufferSize {
		return bufWriter.writeBigChunk(bytes)
	}

	if bytesLen > bufWriter.buffer.Available() {
		n, err = bufWriter.flushInner()
		if err != nil {
			return
		}
	}

	bufWriter.buffer.Write(bytes)

	return len(bytes), nil
}

func (bufWriter *BufferedWriter) Close() error {
	closer, ok :=  bufWriter.innerWriter.(io.Closer)
	if ok {
		return closer.Close()
	}
	
	return nil
}

func (bufWriter *BufferedWriter) Flush() {

	bufWriter.bufferMutex.Lock()
	defer bufWriter.bufferMutex.Unlock()

	bufWriter.flushInner()
}

func (bufWriter *BufferedWriter) flushInner() (n int, err error) {
	bufferedLen := bufWriter.buffer.Buffered()
	flushErr := bufWriter.buffer.Flush()

	return bufWriter.buffer.Buffered() - bufferedLen, flushErr
}

func (bufWriter *BufferedWriter) flushPeriodically() {
	if bufWriter.flushPeriod > 0 {
		ticker := time.NewTicker(bufWriter.flushPeriod)
		for {
			<-ticker.C
			bufWriter.bufferMutex.Lock()
			bufWriter.buffer.Flush()
			bufWriter.bufferMutex.Unlock()
		}
	}
}

func (bufWriter *BufferedWriter) String() string {
	return fmt.Sprintf("BufferedWriter size: %d, flushPeriod: %d", bufWriter.bufferSize, bufWriter.flushPeriod)
}
