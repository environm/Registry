package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
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
	//FileMapping *data.FileMapping
	DataSpecList *data.DataSpecList
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *DeleteHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewDeleteHandler(dataPath string, dataSpecList *data.DataSpecList) *DeleteHandler {
	dh := &DeleteHandler{
		DataPath:     dataPath,
		DataSpecList: dataSpecList,
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

		fileName, tag, _, _, err := utils.GetFileParams(r, "v1.0.0")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting file parameters: %v", err), http.StatusBadRequest)
			return
		}

		//err := d.FileMapping.DeleteFile(fileName, tag, d.DataPath)
		err = d.DataSpecList.DeleteFile(fileName, tag)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete file: %v", err), http.StatusInternalServerError)
			return
		}

		// 删除成功，返回响应
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("File %s (tag: %s) deleted successfully", fileName, tag)))

	}
}
