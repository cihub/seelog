// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package writers

import (
	"testing"
	. "github.com/cihub/seelog/test"
)

func TestChunkWriteOnFilling(t *testing.T) {
	writer, _ := NewBytesVerfier(t)
	bufferedWriter, err := NewBufferedWriter(writer, 1024, 0)

	if err != nil {
		t.Fatalf("Unexpected buffered writer creation error: %s", err.Error())
	}

	bytes := make([]byte, 1000)

	bufferedWriter.Write(bytes)
	writer.ExpectBytes(bytes)
	bufferedWriter.Write(bytes)

	// BufferedWriter writes another chunk not at once but in goroutine (with nondetermined delay)
	// so we wait for a few seconds
	writer.MustNotExpectWithDelay(0.1 * 1e9)
}

func TestFlushByTimePeriod(t *testing.T) {
	writer, _ := NewBytesVerfier(t)
	bufferedWriter, err := NewBufferedWriter(writer, 1024, 100)

	if err != nil {
		t.Fatalf("Unexpected buffered writer creation error: %s", err.Error())
	}

	bytes := []byte("Hello")

	writer.ExpectBytes(bytes)
	bufferedWriter.Write(bytes)
	writer.MustNotExpectWithDelay(0.2 * 1e9)

	// Added after bug with stopped timer
	writer.ExpectBytes(bytes)
	bufferedWriter.Write(bytes)
	writer.MustNotExpectWithDelay(0.2 * 1e9)
}

func TestBigMessageMustPassMemoryBuffer(t *testing.T) {
	writer, _ := NewBytesVerfier(t)
	bufferedWriter, err := NewBufferedWriter(writer, 1024, 0)

	if err != nil {
		t.Fatalf("Unexpected buffered writer creation error: %s", err.Error())
	}

	bytes := make([]byte, 5000)

	for i := 0; i < len(bytes); i++ {
		bytes[i] = uint8(i % 255)
	}

	writer.ExpectBytes(bytes)
	bufferedWriter.Write(bytes)
	writer.MustNotExpectWithDelay(0.1 * 1e9)
}
