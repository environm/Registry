package handler

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"io"
	"net/http"
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
	FileMapping *data.FileMapping
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *DownloadHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewDownloadHandler(dataPath string, fileMapping *data.FileMapping) *DownloadHandler {
	dh := &DownloadHandler{
		DataPath:    dataPath,
		FileMapping: fileMapping,
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

		fileName, tag, err := utils.GetFileParams(r, "v1.0.0")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting file parameters: %v", err), http.StatusBadRequest)
			return
		}

		fileType := r.Header.Get("FileType")
		if fileType == "" {
			fileType = "file"
		}

		clientAddr := r.RemoteAddr

		switch fileType {
		case "folder":
			err := utils.Traverse(d.DataPath, clientAddr)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error traversing directory: %v", err), http.StatusInternalServerError)
				return
			}
		case "file":
			file, err := d.FileMapping.LoadFile(fileName, tag, d.DataPath)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error loading file: %v", err), http.StatusNotFound)
				return
			}
			defer file.Close()
			// 设置响应头，支持文件下载
			w.Header().Set("Content-Type", "application/octet-stream")
			// 将文件内容写入响应
			_, err = io.Copy(w, file)
			file.Close()
			if err != nil {
				http.Error(w, "Failed to send file", http.StatusInternalServerError)
				return
			}
			logs.Infof("File %s (tag: %s) downloaded successfully", fileName, tag)

		default:
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		// 检查文件是否标记为永久存储
		isExits, _ := d.FileMapping.QueryFile(fileName, tag)
		fmt.Printf("IsPermanent: %v", isExits)
		if !isExits.GetIsPermanent() {
			// 文件未标记为永久存储，下载后删除文件
			err := d.FileMapping.DeleteFile(fileName, tag, d.DataPath)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete file after download: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}
}
