package iotest

import (
	"os"
	"syscall"
	"testing"
)

func TestTempFile(t *testing.T) {
	f, cleanup := TempFile(t)
	if _, err := f.Write([]byte("test")); err != nil {
		t.Fatalf("temp file not writable: %v", err)
	}
	cleanup()
	// Confirm closed

	if err := f.Close(); err != syscall.EINVAL {
		t.Errorf("temp file was not closed by cleanup func")
	}
	if _, err := os.Stat(f.Name()); !os.IsNotExist(err) {
		t.Errorf("temp file was not removed by cleanup func")
	}
}

var finfoTests = map[string]string{
	"empty":     "",
	"non-empty": "I am a log file",
}

func TestFileInfo(t *testing.T) {
	for name, in := range finfoTests {
		got := FileInfo(t, []byte(in))
		testEqual(t, name, "size", got.Size(), int64(len(in)))
		testEqual(t, name, "mode", got.Mode(), os.FileMode(0600))
		testEqual(t, name, "isDir", got.IsDir(), false)
	}
}

func testEqual(t *testing.T, name, field string, got, want interface{}) {
	if got != want {
		t.Errorf("%s: incorrect %v: got %q but want %q", name, field, got, want)
	}
}
