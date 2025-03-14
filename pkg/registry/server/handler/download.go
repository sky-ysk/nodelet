package handler

import (
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/utils"
	"io"
	"net/http"
	"path/filepath"
)

// DownloadHandler 对应文件下载请求
// 方法 GET
// URL /download
// Param filename
// TODO:
type DownloadHandler struct {
	//
	DataPath string
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *DownloadHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewDownloadHandler(dataPath string) *DownloadHandler {
	dh := &DownloadHandler{
		DataPath: dataPath,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &DownloadHandler{}

func (d *DownloadHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is supported", http.StatusMethodNotAllowed)
			return
		}
		
		// 获取文件名
		// TODO: 兼容大小写
		fileName := r.URL.Query().Get("filename")
		if fileName == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			return
		}
		fileName = filepath.Clean(fileName)
		
		// 调用文件加载函数
		file, err := utils.LoadFile(d.DataPath, fileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			logs.Infof("Error loading file %s: %v", fileName, err)
			return
		}
		defer file.Close()
		
		// 设置响应头
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		
		// 将文件内容写入响应
		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, "Failed to send file", http.StatusInternalServerError)
			logs.Infof("Error sending file: %v", err)
		}
	}
}
