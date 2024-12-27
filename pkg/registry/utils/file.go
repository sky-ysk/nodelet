package utils

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

func SaveFile(dest, fileName string, data io.Reader) (written int64, err error) {
	// 确保目录存在
	err = os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return 0, err
	}
	// 构造目标文件路径
	filePath := filepath.Join(dest, fileName)
	// 创造目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return 0, err
	}
	defer dst.Close()
	return io.Copy(dst, data)
}

func LoadFile(dest, fileName string) (file *os.File, err error) {
	filePath := filepath.Join(dest, fileName)
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.New("file not found" + fileName)
	}
	file, err = os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
