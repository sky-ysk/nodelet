package utils

import (
	"bytes"
	"errors"
	"sync"
)

// 用于存储文件的线程安全映射
var storage = struct {
	sync.RWMutex
	files map[string]*bytes.Buffer
}{
	files: make(map[string]*bytes.Buffer),
}

// Save 保存文件到内存
func Save(fileName string, data []byte) error {
	storage.Lock()
	defer storage.Unlock()

	storage.files[fileName] = bytes.NewBuffer(data)
	return nil
}

// Load 从内存中读取文件
func Load(fileName string) (*bytes.Buffer, error) {
	storage.RLock()
	defer storage.RUnlock()

	file, exists := storage.files[fileName]
	if !exists {
		return nil, errors.New("file not found")
	}
	return file, nil
}

// Delete 从内存中删除文件
func Delete(fileName string) error {
	storage.Lock()
	defer storage.Unlock()

	delete(storage.files, fileName)
	return nil
}
