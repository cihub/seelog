package writers

import (
	"testing"
)

func TestChunkWriteOnFilling(t *testing.T) {
	testEnv = t

	writer := new(testWriteCloser).Initialize()
	bufferedWriter, err := NewBufferedWriter(writer, 1024, 1, 0)
	
	if err != nil {
		t.Fatalf("Unexpected buffered writer creation error: %s", err.String())
	}
	
	bytes := make([]byte, 1000)

	bufferedWriter.Write(bytes)
	writer.expectBytes(bytes)
	bufferedWriter.Write(bytes)

	// BufferedWriter writes another chunk not at once but in goroutine ( with nondetermined delay )
	//   so we wait few seconds
	writer.mustNotExpectWithDelay(1 * 1e9)
}

func TestFlushByTimePeriod(t *testing.T) {
	testEnv = t

	writer := new(testWriteCloser).Initialize()
	bufferedWriter, err := NewBufferedWriter(writer, 1024, 1, 1)
	
	if err != nil {
		t.Fatalf("Unexpected buffered writer creation error: %s", err.String())
	}
	
	bytes := []byte("Hello")

	writer.expectBytes(bytes)
	bufferedWriter.Write(bytes)
	writer.mustNotExpectWithDelay(2 * 1e9)

	// Added after bug with stopped timer
	writer.expectBytes(bytes)
	bufferedWriter.Write(bytes)
	writer.mustNotExpectWithDelay(2 * 1e9)
}

func TestBigMessageMustPassMemoryBuffer(t *testing.T) {
	testEnv = t

	writer := new(testWriteCloser).Initialize()
	bufferedWriter, err := NewBufferedWriter(writer, 1024, 1, 0)
	
	if err != nil {
		t.Fatalf("Unexpected buffered writer creation error: %s", err.String())
	}
	
	bytes := make([]byte, 1025)

	writer.expectBytes(bytes)
	bufferedWriter.Write(bytes)
	writer.mustNotExpectWithDelay(1 * 1e9)
}
