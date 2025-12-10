package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/mrhyman/gophermart/internal/model"
)

type FileStorage struct {
	mu    sync.RWMutex
	path  string
	orders []model.Order
}

func NewFileStorage(path string) (*FileStorage, error) {
	fs := &FileStorage{
		path:  path,
		orders: make([]model.Order, 0),
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte("[]"), 0644); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &fs.orders); err != nil {
				return nil, err
			}
		}
	}

	return fs, nil
}

func (fs *FileStorage) Ping() error {
	_, err := os.Stat(fs.path)
	return err
}

func (fs *FileStorage) Close() error {
	return nil
}
