package fileManager

import (
	"fmt"

	"hit.edu/framework/pkg/component-base/logs"
	utils "hit.edu/framework/pkg/nodelet/registry/Utils"
)

// 测试使用，暂时只支持单个文件的上传和下载，删除需要手动。
// 在config.go中配置上传和下载的url、保存路径等信息。
const (
	//未下载
	NotDownloaded string = "not_downloaded"
	//正在下载
	Downloading string = "downloading"
	//下载完成
	Downloaded string = "download_completed"
	//下载失败
	DownloadFailed string = "download_failed"
	//未上传
	NotUploaded string = "not_uploaded"
	//正在上传
	Uploading string = "uploading"
	//上传完成
	Uploaded string = "upload_completed"
	//上传失败
	UploadFailed string = "upload_failed"
)
type FileManager struct {
	// 一些配置参数，例如上传和下载的URL、保存路径等
	DataSavedDir string
	UploadURL    string
	ForwardURL   string
	DownloadURL  string
	// 记录本地的runtime的文件下载情况：未下载、正在下载、下载完成、下载失败。key是带后缀的runtime.Name，value是data[]里面所有文件的下载状态
	DownloadStatus map[string]string
	// 记录本地的runtime的文件上传情况：未上传、正在上传、上传完成、上传失败。key是带后缀的runtime.Name，value是data[]里面所有文件的上传状态
	UploadStatus   map[string]string
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
	logs.Infof("DataSavedDir: %s", DataSavedDir)
	logs.Infof("UploadURL: %s", UploadURL)
	logs.Infof("ForwardURL: %s", ForwardURL)
	logs.Infof("DownloadURL: %s", DownloadURL)
	downloadStatus := make(map[string]string)
	uploadStatus := make(map[string]string)
	// 初始化 FileManager
	return &FileManager{
		DataSavedDir: DataSavedDir,
		UploadURL:    UploadURL,
		ForwardURL:   ForwardURL,
		DownloadURL:  DownloadURL,
		DownloadStatus: downloadStatus,
		UploadStatus:   uploadStatus,
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
	logs.Infof("Download url: %s", fm.DownloadURL)
	url := fm.DownloadURL
	downloadURL := url + filename
	fm.DownloadStatus[filename] = Downloading
	// fmt.Println("Download URL:", downloadURL)
	// fmt.Println("Save Path:", savePath)
	err := utils.DownloadFile(downloadURL, savePath)
	if err != nil {
		fmt.Println("Download failed:", err)
		fm.DownloadStatus[filename] = DownloadFailed
		fmt.Printf("file:%s,Download status:%s", filename, DownloadFailed)
		return "", err
	} else {
		fm.DownloadStatus[filename] = Downloaded
		fmt.Println("Download successful!")
		return "Download successful!", nil
	}
	// TODO 更新文件下载的状态 
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
