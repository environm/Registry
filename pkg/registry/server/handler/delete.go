package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
	"path/filepath"
)

// DeleteHandler 对应文件删除请求
// 方法 GET
// URL /delete
// Param filename
// TODO:
type DeleteHandler struct {
	//
	DataPath string
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *DeleteHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewDeleteHandler(dataPath string) *DeleteHandler {
	dh := &DeleteHandler{
		DataPath: dataPath,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &DeleteHandler{}

func (d *DeleteHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is supported", http.StatusMethodNotAllowed)
			return
		}

		// 获取文件名参数
		fileName := r.URL.Query().Get("fileName")
		if fileName == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			return
		}
		fileName = filepath.Clean(fileName) // 清理路径，防止路径遍历攻击

		// 获取标签参数（如果未提供，使用默认值）
		tag := r.URL.Query().Get("tag")
		if tag == "" {
			tag = "v1.0.0"
		}

		// 创建 File 实例
		f := utils.NewFile(fileName, tag)

		// 调用 service 层的 DeleteFile 方法删除文件
		err := f.DeleteFile(d.DataPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete file: %v", err), http.StatusInternalServerError)
			return
		}

		// 删除成功，返回响应
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("File %s (tag: %s) deleted successfully", fileName, tag)))

	}
}
