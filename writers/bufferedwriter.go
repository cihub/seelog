package writers

import (
	"io"
	"os"
	"time"
	"sync"
	"fmt"
	"bufio"
)

// BufferedWriter stores data in memory and flushes it every flushPeriod or when buffer is full
type BufferedWriter struct {
	flushPeriod int           // data flushes interval (in microseconds)
	bufferMutex *sync.Mutex   // mutex for buffer operations syncronization
	innerWriter io.Writer     // inner writer
	buffer      *bufio.Writer // buffered wrapper for inner writer
	bufferSize  int           // max size of data chunk in bytes
}

// NewBufferedWriter creates a new buffered writer struct.
// bufferSize -- size of memory buffer in bytes
// buffersCount -- count of buffers exist at the same time
// flushPeriod -- period in which data flushes from memory buffer in microseconds. 0 - turn off this functionality
func NewBufferedWriter(innerWriter io.Writer, bufferSize int, buffersCount int, flushPeriod int) (*BufferedWriter, os.Error) {

	if innerWriter == nil {
		return nil, os.NewError("Argument is nil: innerWriter")
	}
	if flushPeriod < 0 {
		return nil, os.NewError(fmt.Sprintf("flushPeriod can not be less than 0. Got: %d", buffersCount))
	}

	if bufferSize <= 0 {
		return nil, os.NewError(fmt.Sprintf("bufferSize can not be less or equal to 0. Got: %d", bufferSize))
	}

	buffer, err := bufio.NewWriterSize(innerWriter, bufferSize)

	if err != nil {
		return nil, err
	}

	newWriter := new(BufferedWriter)

	newWriter.innerWriter = innerWriter
	newWriter.buffer = buffer
	newWriter.bufferSize = bufferSize
	newWriter.flushPeriod = flushPeriod
	newWriter.bufferMutex = new(sync.Mutex)

	if flushPeriod != 0 {
		go newWriter.flushPeriodically()
	}

	return newWriter, nil
}

func (bufWriter *BufferedWriter) writeBigChunk(bytes []byte) (n int, err os.Error) {
	bufferedLen := bufWriter.buffer.Buffered()
	flushErr := bufWriter.buffer.Flush()

	if flushErr != nil {
		return bufferedLen - bufWriter.buffer.Buffered(), flushErr
	}

	written, writeErr := bufWriter.innerWriter.Write(bytes)

	if writeErr != nil {
		return bufferedLen + written, writeErr
	}

	return bufferedLen + written, writeErr
}

// Sends data to buffer manager. Waits until all buffers are full.
func (bufWriter *BufferedWriter) Write(bytes []byte) (n int, err os.Error) {
	bufWriter.bufferMutex.Lock()
	defer bufWriter.bufferMutex.Unlock()

	bytesLen := len(bytes)

	if bytesLen > bufWriter.bufferSize {
		return bufWriter.writeBigChunk(bytes)
	}

	if bytesLen > bufWriter.buffer.Available() {
		bufferedLen := bufWriter.buffer.Buffered()
		flushErr := bufWriter.buffer.Flush()

		if flushErr != nil {
			return bufWriter.buffer.Buffered() - bufferedLen, flushErr
		}
	}

	bufWriter.buffer.Write(bytes)

	return len(bytes), nil
}

func (bufWriter *BufferedWriter) flushPeriodically() {
	if bufWriter.flushPeriod > 0 {
		ticker := time.NewTicker(int64(bufWriter.flushPeriod) * 1e6)
		for {
			<-ticker.C
			bufWriter.bufferMutex.Lock()
			bufWriter.buffer.Flush()
			bufWriter.bufferMutex.Unlock()
		}
	}
}
