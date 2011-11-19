// Package writers contains a collection of writers that could be used by sealog dispatchers.
// It allows to write to such receivers as: file, console, rolling(rotation) files, smtp, network, buffered file streams.
package writers

import (
	"io"
	"os"
	"path/filepath"
)

// FileWriter is used to write to a file.
type FileWriter struct {
	innerWriter io.Writer
	fileName    string
}

// Creates a new file and a corresponding writer. Returns error, if the file couldn't be craeted.
func NewFileWriter(fileName string) (writer *FileWriter, err os.Error) {
	newWriter := new(FileWriter)
	
	fileErr := newWriter.createFile()
	if fileErr != nil {
		return nil, fileErr
	}

	return newWriter, nil
}

// Create folder and file on WriteLog/Write first call
func (this *FileWriter) Write(bytes []byte) (n int, err os.Error) {
	return this.innerWriter.Write(bytes)
}

func (this *FileWriter) createFile() os.Error {

	folder, _ := filepath.Split(this.fileName)

	dirErr := fileSystemWrapper.createFolderPath(folder)
	if dirErr != nil {
		return dirErr
	}

	innerWriter, fileError := fileSystemWrapper.create(this.fileName)
	if fileError != nil {
		return fileError
	}

	this.innerWriter = innerWriter

	return nil
}
