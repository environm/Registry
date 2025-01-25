package handler

import (
	"encoding/json"
	"hit.edu/framework/pkg/registry/data"
	"net/http"
)

// QueryIsExistsHandler 对应查询文件是否存在请求
// 方法 GET
// URL /query/list
type QueryListHandler struct {
	//
	DataPath string
	//
	//FileMapping *data.FileMapping
	DataSpecList *data.DataSpecList
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *QueryListHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewQueryListHandler(dataPath string, dataSpecList *data.DataSpecList) *QueryListHandler {
	dh := &QueryListHandler{
		DataPath:     dataPath,
		DataSpecList: dataSpecList,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &QueryListHandler{}

func (d *QueryListHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is supported", http.StatusMethodNotAllowed)
			return
		}
		//lists := d.FileMapping.ListFiles()
		lists := d.DataSpecList.GetList()

		// 将文件列表序列化为 JSON
		response, err := json.Marshal(lists)
		if err != nil {
			http.Error(w, "Failed to serialize file list", http.StatusInternalServerError)
			return
		}

		// 设置响应头并返回结果
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}
}
