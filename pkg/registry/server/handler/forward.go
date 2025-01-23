package handler

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
	"path/filepath"
)

// ForwardHandler 处理 POST 请求并实时转发数据
// TODO:

type ForwardHandler struct {
	//
	DataPath string
	//
	FileMapping *data.FileMapping
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *ForwardHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

// NewForwardHandler 创建一个新的 ForwardHandler 实例
func NewForwardHandler(dataPath string, fileMapping *data.FileMapping) *ForwardHandler {
	dh := &ForwardHandler{
		DataPath:    dataPath,
		FileMapping: fileMapping,
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
		// 获取参数
		param := r.URL.Query()
		// 获取文件名
		fileName := utils.GetQueryParamCaseInsensitive(param, "filename")
		if fileName == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			logs.Infof("Filename missing in request")
			return
		}
		fileName = filepath.Clean(fileName)

		// 获取标签
		tag := utils.GetQueryParamCaseInsensitive(param, "tag")
		if tag == "" {
			tag = "v1.0.0" // 设置默认标签
		}

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
