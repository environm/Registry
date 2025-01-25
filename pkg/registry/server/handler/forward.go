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
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
			return
		}

		fileName, tag, _, _, err := utils.GetFileParams(r, "v1.0.0")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting file parameters: %v", err), http.StatusBadRequest)
			return
		}
		// 获取参数
		param := r.URL.Query()
		// 获取目标地址
		targetURL := utils.GetQueryParamCaseInsensitive(param, "target")
		if targetURL == "" {
			http.Error(w, "Target URL is required for forwarding", http.StatusBadRequest)
			logs.Infof("Target URL missing in request")
			return
		}

		// 在目标 URL 中添加 filename 和 tag 参数
		targetURL = fmt.Sprintf("%s/receive?filename=%s&tag=%s", targetURL, fileName, tag)
		logs.Infof("Forwarding to URL: %s", targetURL)

		// 转发
		if err := utils.ForwardRequest(r, targetURL, w, fileName, tag); err != nil {
			http.Error(w, fmt.Sprintf("Failed to forward request: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
