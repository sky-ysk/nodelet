package handler

import (
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
	"path/filepath"
)

// TODO:

type UploadHandler struct {
	//
	DataPath string
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *UploadHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewUploadHandler(dataPath string) *UploadHandler {
	dh := &UploadHandler{
		DataPath: dataPath,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &UploadHandler{}

func (d *UploadHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
			return
		}
		
		// 获取文件名（从 URL 参数或者 Header 中获取）
		fileName := r.URL.Query().Get("filename")
		if fileName == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			return
		}
		fileName = filepath.Clean(fileName)
		
		//调用文件保存函数
		written, err := utils.SaveFile(d.DataPath, fileName, r.Body)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			logs.Infof("Error saving file: %v", err)
			return
		}
		
		//// 存储到内存
		//buf := new(bytes.Buffer)
		//written, err := io.Copy(buf, r.Body)
		//if err != nil {
		//	http.Error(w, "Failed to read file content", http.StatusInternalServerError)
		//	log.Printf("Error reading file content: %v", err)
		//	return
		//}
		//log.Printf("Received %d bytes from client", written)
		
		// 返回成功响应
		logs.Infof("File %s uploaded successfully, size: %d bytes", fileName, written)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("File uploaded successfully"))
	}
}
