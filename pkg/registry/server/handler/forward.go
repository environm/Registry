package handler

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
)

// ForwardHandler 处理 POST 请求并实时转发数据
// TODO:

type ForwardHandler struct {
	//
	DataPath string
	//
	//FileMapping *data.FileMapping
	DataSpecList *data.DataSpecList
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *ForwardHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

// NewForwardHandler 创建一个新的 ForwardHandler 实例
func NewForwardHandler(dataPath string, dataSpecList *data.DataSpecList) *ForwardHandler {
	dh := &ForwardHandler{
		DataPath:     dataPath,
		DataSpecList: dataSpecList,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &ForwardHandler{}

func (d *ForwardHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取集群 ID
		ClusterID := r.Header.Get("ClusterID")
		if ClusterID == "" {
			http.Error(w, "Missing ClusterID in request header", http.StatusBadRequest)
			logs.Infof("Missing ClusterID in request header")
			return
		}

		// 生成新的 URL
		//newURL, err := utils.TransformURL(ClusterID, r)
		newURL, err := utils.TransformURL(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logs.Infof("Forwarding to URL: %s", newURL)

		// 转发
		if err := utils.ForwardRequest(r, newURL, w); err != nil {
			http.Error(w, fmt.Sprintf("Failed to forward request: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
