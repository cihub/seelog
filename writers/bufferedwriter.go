package writers

import (
	"io"
	"os"
	"time"
)

// BufferedWriter stores data in memory and flushes it every flushPeriod or when buffer is full
type BufferedWriter struct {
	innerWriter  io.Writer
	bufferSize   int // size of memory buffer in bytes
	buffersCount int // count of buffers exist at the same time
	flushPeriod  int // period in witch data flushes from memory buffer in seconds

	forceFlush    chan int     // send signal to force flush
	emptyBuffers  chan *buffer // next buffer retrieved when previous get full
	writeToBuffer chan []byte

	flushSync chan int
}

// NewBufferedWriter creates a new buffered writer struct.
//  bufferSize - size of memory buffer in bytes
//  buffersCount - count of buffers exist at the same time
//  flushPeriod - period in which data flushes from memory buffer in seconds. 0 - turn off this functionality
func NewBufferedWriter(innerWriter io.Writer, bufferSize int, buffersCount int, flushPeriod int) (*BufferedWriter, os.Error) {
	
	if innerWriter == nil {
		return nil, os.NewError("BufferedWriter.Init Argument nil: innerWriter")
	}
	if buffersCount <= 0 {
		return nil, os.NewError("BufferedWriter.Init buffersCount can not be less then 1")
	}
	if bufferSize < 1024 {
		return nil, os.NewError("BufferedWriter.Init bufferSize can not be less then 1kB")
	}
	if flushPeriod < 0 {
		return nil, os.NewError("BufferedWriter.Init flushPeriod can not be less then 0")
	}

	newWriter := new(BufferedWriter)

	newWriter.innerWriter = innerWriter
	newWriter.bufferSize = bufferSize
	newWriter.buffersCount = buffersCount
	newWriter.flushPeriod = flushPeriod

	newWriter.emptyBuffers = make(chan *buffer, buffersCount)
	newWriter.forceFlush = make(chan int)
	newWriter.writeToBuffer = make(chan []byte)

	newWriter.flushSync = make(chan int, 1)
	newWriter.flushSync <- 1

	go newWriter.manageBuffers()

	if flushPeriod != 0 {
		go newWriter.flushPeriodically()
	}

	return newWriter, nil
}

// Sends data to buffer manager
// Will wait if all buffers are full
func (this *BufferedWriter) Write(bytes []byte) (n int, err os.Error) {
	this.writeToBuffer <- bytes

	return len(bytes), nil
}

// In infinite loops waitd following events:
//   1. Flush signal - flushd current buffer
//   2. Data to write - writes to buffer untill it is full, then flushes buffer and writes rest data to next buffer
//      If no one buffer available waits.
func (this *BufferedWriter) manageBuffers() {
	// Create all buffers
	for i := 0; i < this.buffersCount; i++ {
		this.emptyBuffers <- new(buffer).Initialize(this.bufferSize)
	}

	currentBuffer := <-this.emptyBuffers
	for {
		select {
		case <-this.forceFlush:
			go this.flush(currentBuffer)
			currentBuffer = <-this.emptyBuffers
		case bytes := <-this.writeToBuffer:
			// Too big data goes straight to writer
			if len(bytes) >= this.bufferSize {
				go this.flushWithBigChunk(currentBuffer, bytes)
				currentBuffer = <-this.emptyBuffers
			}

			if !currentBuffer.Write(bytes) {
				go this.flush(currentBuffer)

				currentBuffer = <-this.emptyBuffers
				currentBuffer.Write(bytes)
			}

			if currentBuffer.isFull() {
				go this.flush(currentBuffer)
				currentBuffer = <-this.emptyBuffers
			}
		}
	}
}

func (this *BufferedWriter) flushPeriodically() {
	ticker := time.NewTicker(int64(this.flushPeriod) * 1e9)
	for {
		<-ticker.C
		this.flushCurrent()
	}
}

func (this *BufferedWriter) flushCurrent() {
	this.forceFlush <- 1
}

func (this *BufferedWriter) flush(buffer *buffer) {
	<-this.flushSync

	this.flushBuffer(buffer)

	this.flushSync <- 1
}

func (this *BufferedWriter) flushWithBigChunk(buffer *buffer, bytes []byte) {
	<-this.flushSync

	this.flushBuffer(buffer)
	this.innerWriter.Write(bytes)

	this.flushSync <- 1
}

func (this *BufferedWriter) flushBuffer(buffer *buffer) {
	if !buffer.isEmpty() {
		buffer.flush(this.innerWriter)
	}
	this.emptyBuffers <- buffer
}
