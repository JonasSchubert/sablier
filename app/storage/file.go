package storage

import (
	"fmt"
	"io"
	"os"

	"github.com/sablierapp/sablier/config"
	log "github.com/sirupsen/logrus"
)

type Storage interface {
	Reader() (io.ReadCloser, error)
	Writer() (io.WriteCloser, error)

	Enabled() bool
}

type FileStorage struct {
	file string
}

func NewFileStorage(config config.Storage) (Storage, error) {
	storage := &FileStorage{
		file: config.File,
	}

	if storage.Enabled() {
		file, err := os.OpenFile(config.File, os.O_RDWR|os.O_CREATE, 0755)

		if err != nil {
			return nil, err
		}
		defer file.Close()

		stats, err := file.Stat()
		if err != nil {
			return nil, err
		}

		// Initialize file to an empty JSON3
		if stats.Size() == 0 {
			file.WriteString("{}")
		}

		log.Infof("initialized storage to %s", config.File)
	} else {
		log.Warn("no storage configuration provided. all states will be lost upon exit")
	}
	return storage, nil
}

func (fs *FileStorage) Reader() (io.ReadCloser, error) {
	if !fs.Enabled() {
		return nil, fmt.Errorf("file storage is not enabled")
	}
	return os.OpenFile(fs.file, os.O_RDWR|os.O_CREATE, 0755)
}

func (fs *FileStorage) Writer() (io.WriteCloser, error) {
	if !fs.Enabled() {
		return nil, fmt.Errorf("file storage is not enabled")
	}
	return os.OpenFile(fs.file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
}

func (fs *FileStorage) Enabled() bool {
	return len(fs.file) > 0
}
