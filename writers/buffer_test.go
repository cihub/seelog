package writers

import (
	"testing"
)

func TestBuffer_Write(t *testing.T) {
	testEnv = t

	writer := new(testWriteCloser).Initialize()
	buf := new(buffer).Initialize(1024)

	bytes := []byte("Hello")
	buf.Write(bytes)

	writer.expectBytes(bytes)
	buf.flush(writer)
	writer.mustNotExpect()
}

func TestBuffer_Clear(t *testing.T) {
	testEnv = t

	writer := new(testWriteCloser).Initialize()
	buf := new(buffer).Initialize(1024)

	bytes := []byte("Hello")
	buf.Write(bytes)
	buf.clear()

	writer.expectBytes([]byte(""))
	buf.flush(writer)
	writer.mustNotExpect()
}

func TestBuffer_isEmpty(t *testing.T) {
	testEnv = t

	writer := new(emptyWriteCloser).Initialize()
	buf := new(buffer).Initialize(1024)

	if !buf.isEmpty() {
		testEnv.Errorf("Buffer is not empty after creation")
	}

	buf.Write([]byte(""))
	if !buf.isEmpty() {
		testEnv.Errorf("Buffer is not empty after empty write")
	}

	buf.Write([]byte("Hello"))
	if buf.isEmpty() {
		testEnv.Errorf("Buffer is empty after write")
	}

	buf.clear()
	if !buf.isEmpty() {
		testEnv.Errorf("Buffer is not empty after clear")
	}

	buf.Write([]byte("Hello"))
	buf.flush(writer)
	if !buf.isEmpty() {
		testEnv.Errorf("Buffer is not empty after flush")
	}
}

func TestBuffer_isFull(t *testing.T) {
	testEnv = t

	bytes := []byte("Hello")

	buf := new(buffer).Initialize(len(bytes))

	buf.Write(bytes)
	if !buf.isFull() {
		testEnv.Errorf("Buffer is not full")
	}
}
