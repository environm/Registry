package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
	"path/filepath"
)

// QueryIsExistsHandler 对应查询文件是否存在请求
// 方法 GET
// URL /query/exits
// Param filename
// TODO:
type QueryIsExistsHandler struct {
	//
	DataPath string
	//
	FileMapping *utils.FileMapping
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *QueryIsExistsHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewQueryIsExistsHandler(dataPath string, fileMapping *utils.FileMapping) *QueryIsExistsHandler {
	dh := &QueryIsExistsHandler{
		DataPath:    dataPath,
		FileMapping: fileMapping,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &QueryIsExistsHandler{}

func (d *QueryIsExistsHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
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

		// 创建 File 实例并调用 IsExists 检查文件是否存在
		//f := utils.NewFile(fileName, tag)
		//exists := f.IsExists(d.DataPath)
		exists, err := d.FileMapping.QueryFile(fileName, tag)
		if err != nil {
			http.Error(w, fmt.Sprintf("查询文件失败: %v", err), http.StatusInternalServerError)
			return
		}

		if exists != nil {
			// 文件存在，返回 200 OK
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("File %s (tag: %s) exists", fileName, tag)))
		} else {
			// 文件不存在，返回 404 Not Found
			http.Error(w, fmt.Sprintf("File %s (tag: %s) not found", fileName, tag), http.StatusNotFound)
		}

	}
}
