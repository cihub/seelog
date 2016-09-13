package seelog

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestGzip(t *testing.T) {
	defer cleanupWriterTest(t)

	files := make(map[string][]byte)
	files["file1"] = []byte("I am a log")

	readers := make(map[string]io.Reader)
	for fname, fcont := range files {
		readers[fname] = bytes.NewReader(fcont)
	}

	err := createGzip("./gzip.gz", readers["file1"])
	if err != nil {
		t.Fatal(err)
	}

	decompressedFile, err := unGzip("./gzip.gz")
	if err != nil {
		t.Fatal(err)
	}

	equal := reflect.DeepEqual(files["file1"], decompressedFile)
	if !equal {
		t.Fatal("gzip(ungzip(file)) should be equal to file")
	}
}

func TestTar(t *testing.T) {
	defer cleanupWriterTest(t)
	files := make(map[string][]byte)
	files["file1"] = []byte("I am a log")
	files["file2"] = []byte("I am another log")

	readers := make(map[string]io.Reader)
	for fname, fcont := range files {
		readers[fname] = bytes.NewReader(fcont)
	}

	tar, err := createTar(readers)
	if err != nil {
		t.Fatal(err)
	}

	resultFiles, err := unTar(tar)
	if err != nil {
		t.Fatal(err)
	}
	equal := reflect.DeepEqual(files, resultFiles)
	if !equal {
		t.Fatal("untar(tar(files)) should be equal to files")
	}
}

func TestIsTar(t *testing.T) {
	defer cleanupWriterTest(t)
	files := make(map[string][]byte)
	files["file1"] = []byte("I am a log")
	files["file2"] = []byte("I am another log")

	readers := make(map[string]io.Reader)
	for fname, fcont := range files {
		readers[fname] = bytes.NewReader(fcont)
	}

	tar, _ := createTar(readers)

	if !isTar(tar) {
		t.Fatal("tar(files) should be recognized as a tar file")
	}
}
