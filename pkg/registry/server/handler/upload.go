package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
	"path/filepath"
)

// TODO:

type UploadHandler struct {
	//
	DataPath string
	//
	FileMapping *data.FileMapping
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *UploadHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewUploadHandler(dataPath string, fileMapping *data.FileMapping) *UploadHandler {
	dh := &UploadHandler{
		DataPath:    dataPath,
		FileMapping: fileMapping,
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
		// 获取参数
		param := r.URL.Query()
		// 获取文件名（从 URL 参数或者 Header 中获取）
		fileName := utils.GetQueryParamCaseInsensitive(param, "filename")
		if fileName == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			return
		}
		fileName = filepath.Clean(fileName)

		// 获取标签（如果没有提供，则设置默认标签）
		tag := utils.GetQueryParamCaseInsensitive(param, "tag")
		if tag == "" {
			tag = "v1.0.0" // 默认标签
		}

		fileType := r.Header.Get("FileType")
		if fileType == "" {
			fileType = "file"
			//http.Error(w, "FileType header is required", http.StatusBadRequest)
		}

		// 获取是否需要永久存储的参数
		//isPermanent := utils.GetQueryParamCaseInsensitive(param, "isPermanent") == "true"

		// 通过 upload 上传的文件默认为持久化存储， receive 收到的文件默认为临时存储(区别在于 下载之后是否删除)

		switch fileType {
		case "folder", "completion":
			utils.ReceiveDir(w, r, d.DataPath)
			err := d.FileMapping.SaveFolder(d.DataPath, fileName, tag, true)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to save folder: %v", err), http.StatusInternalServerError)
				return
			}
		case "file":
			err := d.FileMapping.SaveFile(d.DataPath, fileName, tag, true, r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("File uploaded successfully"))
	}
}
