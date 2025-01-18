package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/data"
	"net/http"
	"os"
	"path/filepath"
)

// TODO:

type CatalogueDownloadHandler struct {
	//
	DataPath string
	//
	FileMapping *data.FileMapping
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *CatalogueDownloadHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewCatalogueDownloadHandler(dataPath string, fileMapping *data.FileMapping) *CatalogueDownloadHandler {
	dh := &CatalogueDownloadHandler{
		DataPath:    dataPath,
		FileMapping: fileMapping,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &CatalogueDownloadHandler{}

func (d *CatalogueDownloadHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is supported", http.StatusMethodNotAllowed)
			return
		}

		// 从查询参数中获取 baseDir
		param := r.URL.Query()
		baseDir := param.Get("baseDir")
		if baseDir == "" {
			http.Error(w, "baseDir parameter is required", http.StatusBadRequest)
			return
		}

		// 解析 baseDir 的路径
		cleanPath := filepath.Clean(baseDir)
		// 如果是相对路径，则拼接 DataPath
		if !filepath.IsAbs(cleanPath) {
			cleanPath = filepath.Join(d.DataPath, cleanPath)
		}

		fullPath := cleanPath
		if fullPath == "" {
			http.Error(w, "Invalid baseDir parameter", http.StatusBadRequest)
			return
		}

		// 检查路径是否在 DataPath 范围内
		if !filepath.HasPrefix(fullPath, filepath.Clean(d.DataPath)) {
			http.Error(w, "Invalid baseDir parameter", http.StatusForbidden)
			return
		}

		// 检查路径是否存在并确保是目录
		info, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Path %s does not exist", baseDir), http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, fmt.Sprintf("Error accessing path: %v", err), http.StatusInternalServerError)
			return
		}
		if !info.IsDir() {
			http.Error(w, "The specified path is not a directory", http.StatusBadRequest)
			return
		}

		// 压缩并下载目录
		err = d.FileMapping.ZipAndDownload(w, fullPath, filepath.Base(fullPath))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to download directory: %v", err), http.StatusInternalServerError)
		}

	}
}
