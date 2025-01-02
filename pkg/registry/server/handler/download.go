package handler

import (
	"fmt"
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
	FileMapping *utils.FileMapping
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *DownloadHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewDownloadHandler(dataPath string, fileMapping *utils.FileMapping) *DownloadHandler {
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
		// 获取参数
		param := r.URL.Query()
		// 获取文件名
		//fileName := r.URL.Query().Get("filename")
		fileName := utils.GetQueryParamCaseInsensitive(param, "filename")
		if fileName == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			return
		}
		fileName = filepath.Clean(fileName)

		// 从 URL 查询参数中获取文件标签（可选）
		//tag := r.URL.Query().Get("tag")
		tag := utils.GetQueryParamCaseInsensitive(param, "tag")
		if tag == "" {
			tag = "v1.0.0" // 默认标签
		}

		// 创建 File 实例
		//f := utils.NewFile(fileName, tag)

		// 调用 LoadFile 加载文件
		//file, err := f.LoadFile(d.DataPath)
		file, err := d.FileMapping.LoadFile(fileName, tag, d.DataPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error loading file: %v", err), http.StatusNotFound)
			return
		}
		defer file.Close()

		// 设置响应头，支持文件下载
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

		// 将文件内容写入响应
		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, "Failed to send file", http.StatusInternalServerError)
			return
		}
		logs.Infof("File %s (tag: %s) downloaded successfully", fileName, tag)

		// 检查文件是否标记为永久存储
		//isExits, _ := utils.GetIsPermanent(d.DataPath, fileName, tag)
		isExits, _ := d.FileMapping.QueryFile(fileName, tag)
		fmt.Printf("IsPermanent: %v", isExits)
		if !isExits.GetIsPermanent() {
			//if isExits, _ := utils.GetIsPermanent(d.DataPath, fileName, tag); isExits {
			// 文件未标记为永久存储，下载后删除文件
			//err := f.DeleteFile(d.DataPath)
			err := d.FileMapping.DeleteFile(fileName, tag, d.DataPath)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete file after download: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}
}
