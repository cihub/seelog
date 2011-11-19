package writers

import (
	"io"
)

// Represents array of bytes with fill percentage tracking
type buffer struct {
	bytes      []byte
	length     int
	freeLength int
}

func (this *buffer) Initialize(length int) *buffer {
	this.bytes = make([]byte, length)
	this.length = length
	this.freeLength = this.length

	return this
}

func (this *buffer) Write(bytes []byte) bool {
	if this.freeLength < len(bytes) {
		return false
	}

	startCopyFrom := this.length - this.freeLength
	copy(this.bytes[startCopyFrom:startCopyFrom+len(bytes)], bytes)
	this.freeLength -= len(bytes)
	return true
}

func (this *buffer) isEmpty() bool {
	return this.freeLength == this.length
}

func (this *buffer) isFull() bool {
	return this.freeLength == 0
}

func (this *buffer) flush(writer io.Writer) {
	writer.Write(this.bytes[0 : this.length-this.freeLength])
	this.freeLength = this.length
}

func (this *buffer) clear() {
	this.freeLength = this.length
}
