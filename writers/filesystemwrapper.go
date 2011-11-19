package writers

// fileSystemWrapperInterface is used for testing. When sealog is used in a real app, osWrapper uses standard os
// funcs. When sealog is being tested, fileSystemWrapperTest emulates some of the os funcs.

import (
	"os"
	"io"
)

const (
	defaultFilePermissions      = 0666
	defaultDirectoryPermissions = 0766
)

var fileSystemWrapper fileSystemWrapperInterface = new(osWrapper)

type fileSystemWrapperInterface interface {
	createFolderPath(folderPath string) os.Error
	create(fileName string) (io.WriteCloser, os.Error)
	getFileSize(fileName string) (int64, os.Error)
	getFileNames(folderPath string) ([]string, os.Error)
	rename(fileNameFrom string, fileNameTo string) os.Error
	remove(fileName string) os.Error
}

type osWrapper struct {

}

func (this *osWrapper) createFolderPath(folderPath string) os.Error {
	if folderPath == "" {
		return nil
	}

	_, dirErr := os.Lstat(folderPath)
	if dirErr != nil {
		dirErr = os.MkdirAll(folderPath, defaultDirectoryPermissions)
		if dirErr != nil {
			return dirErr
		}
	}

	return nil
}
func (this *osWrapper) create(fileName string) (io.WriteCloser, os.Error) {
	return os.Create(fileName)
}
func (this *osWrapper) getFileSize(fileName string) (int64, os.Error) {
	stat, err := os.Lstat(fileName)
	if err != nil {
		return 0, err
	}

	return stat.Size, nil
}
func (this *osWrapper) getFileNames(folderPath string) ([]string, os.Error) {
	if folderPath == "" {
		folderPath = "."
	}

	folder, err := os.Open(folderPath)
	if err != nil {
		return make([]string, 0), err
	}

	files, err := folder.Readdirnames(-1)
	if err != nil {
		return make([]string, 0), err
	}

	return files, nil
}
func (this *osWrapper) rename(fileNameFrom string, fileNameTo string) os.Error {
	return os.Rename(fileNameFrom, fileNameTo)
}
func (this *osWrapper) remove(fileName string) os.Error {
	return os.Remove(fileName)
}
