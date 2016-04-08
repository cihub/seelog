package seelog

import (
	"testing"
	"fmt"
	"reflect"
)


func TestGzip(t *testing.T) {
	defer cleanupWriterTest(t)

	files := make(map[string][]byte)
	files["file1"] = []byte("I am a log")
	err := createGzip("./gzip.gz", files["file1"])
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	decompressedFile, err := unGzip("./gzip.gz")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	equal := reflect.DeepEqual(files["file1"], decompressedFile)
	if !equal {
		t.Fail()
	}
}

func TestTar(t *testing.T) {
	defer cleanupWriterTest(t)
	files := make(map[string][]byte)
	files["file1"] = []byte("I am a log")
	files["file2"] = []byte("I am another log")
	tar, _ := createTar(files)

	resultFiles, _ := unTar(tar)
	equal := reflect.DeepEqual(files, resultFiles)
	if !equal {
		t.Fail()
	}
}
