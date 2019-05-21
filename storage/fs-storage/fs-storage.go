package fs_storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"univer/twitter-crawler/log"
)

const rootFolderPath = "data"

type FsStorage struct {
	*log.Logger
}

func NewFsStorage() *FsStorage {
	logger := log.NewLogger("FsStorage")

	stor := &FsStorage{}
	stor.Logger = logger
	return stor
}

func (stor *FsStorage) Save(record *Record) error {
	dirPath := path.Join(rootFolderPath, record.Entity, record.Id)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, os.ModePerm)
	}

	files,_ := ioutil.ReadDir(dirPath)
	nextId := len(files)

	fileName := path.Join(dirPath, fmt.Sprintf("%d.%s", nextId, record.Format))
	err := ioutil.WriteFile(fileName, record.Data, os.ModePerm)

	return err
}

func (stor *FsStorage) GetLastCursor() (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (stor *FsStorage) SetLastCursor(cursor string) (error) {
	return fmt.Errorf("not implemented")
}