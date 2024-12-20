package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
	"path/filepath"
)

// TODO:

type UploadHandler struct {
	//
	DataPath string
	//
	FileMapping *utils.FileMapping
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *UploadHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewUploadHandler(dataPath string, fileMapping *utils.FileMapping) *UploadHandler {
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

		// 获取文件名（从 URL 参数或者 Header 中获取）
		fileName := r.URL.Query().Get("filename")
		if fileName == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			return
		}
		fileName = filepath.Clean(fileName)

		// 获取标签（如果没有提供，则设置默认标签）
		tag := r.URL.Query().Get("tag")
		if tag == "" {
			tag = "v1.0.0" // 默认标签
		}

		// 获取是否需要永久存储的参数（默认为 false）
		isPermanent := r.URL.Query().Get("isPermanent") == "true"

		//isPermanent := false // 默认值
		//if isPermanentStr != "" {
		//	// 将字符串转换为布尔值
		//	var err error
		//	isPermanent, err = strconv.ParseBool(isPermanentStr)
		//	if err != nil {
		//		http.Error(w, "Invalid value for isPermanent (must be true or false)", http.StatusBadRequest)
		//		return
		//	}
		//}

		// 获取上传的文件
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "无法获取文件", http.StatusBadRequest)
			return
		}
		defer file.Close()

		//// 创建 File 实例
		//f := utils.NewFile(fileName, tag, isPermanent)
		err = d.FileMapping.SaveFile(d.DataPath, fileName, tag, isPermanent, file)
		if err != nil {
			http.Error(w, fmt.Sprintf("保存文件失败: %v", err), http.StatusInternalServerError)
			return
		}

		// 根据 isPermanent 参数设置文件是否为永久存储
		//f.SetPermanent(isPermanent)

		// 使用服务层的 SaveFile 方法保存文件
		//err = f.SaveFile(d.DataPath, file)
		//if err != nil {
		//	http.Error(w, fmt.Sprintf("保存文件失败: %v", err), http.StatusInternalServerError)
		//	return
		//}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("File uploaded successfully"))
	}
}
