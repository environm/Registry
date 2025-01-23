package handler

import (
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/registry/data"
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
	FileMapping *data.FileMapping
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *QueryIsExistsHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewQueryIsExistsHandler(dataPath string, fileMapping *data.FileMapping) *QueryIsExistsHandler {
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
		// 获取参数
		param := r.URL.Query()

		// 获取文件名参数
		fileName := utils.GetQueryParamCaseInsensitive(param, "filename")
		if fileName == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			return
		}
		fileName = filepath.Clean(fileName) // 清理路径，防止路径遍历攻击

		// 获取标签参数（如果未提供，使用默认值）
		tag := utils.GetQueryParamCaseInsensitive(param, "tag")
		if tag == "" {
			tag = "v1.0.0"
		}

		// 创建 File 实例并调用 IsExists 检查文件是否存在
		file, err := d.FileMapping.QueryFile(fileName, tag)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to check file existence: %v", err), http.StatusInternalServerError)
			return
		}

		if file != nil {
			// 文件存在，返回 200 OK
			//w.WriteHeader(http.StatusOK)
			//w.Write([]byte(fmt.Sprintf("File %s (tag: %s) exists", fileName, tag)))
			// 文件存在，返回文件的 JSON 内容
			w.Header().Set("Content-Type", "application/json")
			jsonData, err := json.Marshal(file)
			if err != nil {
				http.Error(w, "Failed to marshal file data", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(jsonData)
		} else {
			// 文件不存在，返回 404 Not Found
			http.Error(w, fmt.Sprintf("File %s (tag: %s) not found", fileName, tag), http.StatusNotFound)
		}

	}
}
