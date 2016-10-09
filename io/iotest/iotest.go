package iotest

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

// TempFile creates a new temporary file for testing and returns the file
// pointer and a cleanup function for closing and removing the file.
func TempFile(t *testing.T) (*os.File, func()) {
	f, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	return f, func() {
		os.Remove(f.Name())
		f.Close()
	}
}

// FileInfo computes the os.FileInfo for a given byte slice.
func FileInfo(t *testing.T, fbytes []byte) os.FileInfo {
	// Get FileInfo
	f, clean := TempFile(t)
	defer clean()
	_, err := io.Copy(f, bytes.NewReader(fbytes))
	if err != nil {
		t.Fatalf("copy to temp file: %v", err)
	}
	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("stat temp file: %v", err)
	}
	return fi
}
