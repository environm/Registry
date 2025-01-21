package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/data"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// TODO:

type CatalogueUploadHandler struct {
	//
	DataPath string
	//
	FileMapping *data.FileMapping
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *CatalogueUploadHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewCatalogueUploadHandler(dataPath string, fileMapping *data.FileMapping) *CatalogueUploadHandler {
	dh := &CatalogueUploadHandler{
		DataPath:    dataPath,
		FileMapping: fileMapping,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &CatalogueUploadHandler{}

func (d *CatalogueUploadHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
			return
		}

		contentType := r.Header.Get("Content-Type")
		rootPath := r.Header.Get("rootPath") // 根目录路径
		parentPath := r.Header.Get("Parent-Path")

		if rootPath == "" || parentPath == "" {
			http.Error(w, "Root path or parent path is missing", http.StatusBadRequest)
			return
		}

		rootPath = filepath.Clean(rootPath)
		parentPath = filepath.Clean(parentPath)

		if contentType == "application/text" {
			// 处理目录
			dirName := r.Header.Get("Directory-Name")
			if dirName == "" {
				http.Error(w, "Directory name is missing", http.StatusBadRequest)
				return
			}
			// 构建目标路径
			dirPath := filepath.Join(d.DataPath, parentPath, dirName)

			err := d.FileMapping.CreateDir(dirPath)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to create directory: %v", err), http.StatusInternalServerError)
				return
			}
			w.Write([]byte("Directory created successfully"))
		} else if contentType == "application/octet-stream" {
			// 处理文件
			fileName := r.Header.Get("File-Name")
			if fileName == "" {
				http.Error(w, "File name is missing", http.StatusBadRequest)
				return
			}

			// 构建目标路径
			filePath := filepath.Join(d.DataPath, parentPath, fileName)
			parentFullPath := filepath.Dir(filePath)

			err := d.FileMapping.CreateDir(parentFullPath) // 确保父目录存在
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to create parent directory: %v", err), http.StatusInternalServerError)
				return
			}

			// 创建并写入文件
			file, err := os.Create(filePath)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to create file: %v", err), http.StatusInternalServerError)
				return
			}
			defer file.Close()

			_, err = io.Copy(file, r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to write file: %v", err), http.StatusInternalServerError)
				return
			}

			w.Write([]byte("File uploaded successfully"))
		} else {
			http.Error(w, "Unsupported content type", http.StatusBadRequest)
		}
	}
}
