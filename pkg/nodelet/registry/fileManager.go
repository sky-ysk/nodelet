package fileManager

import (
	"fmt"

	utils "hit.edu/framework/pkg/nodelet/registry/Utils"
)

// 测试使用，暂时只支持单个文件的上传和下载，删除需要手动。
// 在config.go中配置上传和下载的url、保存路径等信息。

type FileManager struct {
	DataSavedDir string
	UploadURL    string
	ForwardURL   string
	DownloadURL string
}

// type FileHandler interface {
// 	UploadFile(filePath string) (string, error)
// 	uploadDir(dirPath string) (string, error)
// 	DownloadFile(filename, savePath string) (string, error)
// 	DownloadDir(dirPath, savePath string) (string, error)
// 	ForwardFile(filePath string) (string, error)
// 	DeleteFile(filePath string) (string, error)
// 	DeleteDir(dirPath string) (string, error)
// }

func NewFileManager() *FileManager {
	return &FileManager{
		DataSavedDir: DataSavedDir,
		UploadURL:    UploadURL,
		ForwardURL:   ForwardURL,
		DownloadURL:  DownloadURL,
	}
}

func (fm *FileManager) UploadFile(filePath string) (string, error) {
	// 调用 utils.UploadFile 函数上传文件
	// 这里的 filePath 是要上传的文件路径
	// 返回上传结果和错误信息
	url := fm.UploadURL
	err := utils.UploadFile(filePath, "v1.0.0", url)
	if err != nil {
		fmt.Println("Upload failed:", err)
		return "", err
	} else {
		fmt.Println("Upload successful!")
		return "Upload successful!", nil
	}
}

// func (fm *FileManager) uploadDir(dirPath string) (string, error) {
// 	// 调用 utils.UploadDir 函数上传目录
// 	// 这里的 dirPath 是要上传的目录路径
// 	// 返回上传结果和错误信息
// 	return utils.UploadDir(dirPath, fm.UploadURL)
// }

func (fm *FileManager) DownloadFile(filename, savePath string) (string, error) {
	// 调用 utils.DownloadFile 函数下载文件
	// 这里的 filename 是要下载的文件名，savePath 是保存路径
	// 返回下载结果和错误信息
	url := fm.DownloadURL
	downloadURL := url + filename
	// fmt.Println("Download URL:", downloadURL)
	// fmt.Println("Save Path:", savePath)
	err := utils.DownloadFile(downloadURL, savePath)
	if err != nil {
		fmt.Println("Download failed:", err)
		return "", err
	} else {
		fmt.Println("Download successful!")
		return "Download successful!", nil
	}
}

// func (fm *FileManager) DownloadDir(dirPath, savePath string) (string, error) {
// 	// 调用 utils.DownloadDir 函数下载目录
// 	// 这里的 dirPath 是要下载的目录路径，savePath 是保存路径
// 	// 返回下载结果和错误信息
// 	return utils.DownloadDir(dirPath, savePath)
// }

// func (fm *FileManager) ForwardFile(filePath string) (string, error) {
// 	// 调用 utils.ForwardFile 函数转发文件
// 	// 这里的 filePath 是要转发的文件路径
// 	// 返回转发结果和错误信息
// 	return utils.ForwardFile(filePath, fm.ForwardURL)
// }

// func (fm *FileManager) DeleteFile(filePath string) (string, error) {
// 	// 调用 utils.DeleteFile 函数删除文件
// 	// 这里的 filePath 是要删除的文件路径
// 	// 返回删除结果和错误信息
// 	return utils.DeleteFile(filePath)
// }

// func (fm *FileManager) DeleteDir(dirPath string) (string, error) {
// 	// 调用 utils.DeleteDir 函数删除目录
// 	// 这里的 dirPath 是要删除的目录路径
// 	// 返回删除结果和错误信息
// 	return utils.DeleteDir(dirPath)
// }

// func (fm *FileManager) Traverse(filePath string) (string, error) {
// 	// 调用 utils.Traverse 函数遍历目录
// 	// 这里的 filePath 是要遍历的目录路径
// 	// 返回遍历结果和错误信息
// 	return utils.Traverse(filePath, fm.UploadURL)
// }
