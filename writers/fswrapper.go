package writers

// Real FileSystemWrapperInterface implementation that uses os package.

import (
	"os"
	"io"
	"github.com/cihub/sealog/test"
)

const (
	defaultFilePermissions      = 0666
	defaultDirectoryPermissions = 0767
)

var fileSystemWrapper test.FileSystemWrapperInterface = new(osWrapper)

// SetTestMode is used for testing purposes only! Do not use that or you may get incorrect behavior
func SetTestMode(testWrapper test.FileSystemWrapperInterface) {
	fileSystemWrapper = testWrapper
}

type osWrapper struct {

}

func (_ *osWrapper) MkdirAll(folderPath string) os.Error {
	if folderPath == "" {
		return nil
	}

	return os.MkdirAll(folderPath, defaultDirectoryPermissions)
}
func (_ *osWrapper) Create(fileName string) (io.WriteCloser, os.Error) {
	return os.Create(fileName)
}
func (_ *osWrapper) GetFileSize(fileName string) (int64, os.Error) {
	stat, err := os.Lstat(fileName)
	if err != nil {
		return 0, err
	}

	return stat.Size, nil
}
func (_ *osWrapper) GetFileNames(folderPath string) ([]string, os.Error) {
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
func (_ *osWrapper) Rename(fileNameFrom string, fileNameTo string) os.Error {
	return os.Rename(fileNameFrom, fileNameTo)
}
func (_ *osWrapper) Remove(fileName string) os.Error {
	return os.Remove(fileName)
}
func (_ *osWrapper) Exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
