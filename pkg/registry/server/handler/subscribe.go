package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
	"path/filepath"
)

// SubscribeHandler 文件数据订阅
// 方法 Post
// URL /subscribe
// Param filename tag
// TODO:
type SubscribeHandler struct {
	//
	Subscribers *data.SubscriptionManager
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *SubscribeHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewSubscribeHandler(subscribers *data.SubscriptionManager) *SubscribeHandler {
	dh := &SubscribeHandler{
		Subscribers: subscribers,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &SubscribeHandler{}

func (d *SubscribeHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取参数
		param := r.URL.Query()
		// 获取文件名
		fileName := utils.GetQueryParamCaseInsensitive(param, "filename")
		if fileName == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			return
		}
		fileName = filepath.Clean(fileName)

		// 获取 tag
		tag := utils.GetQueryParamCaseInsensitive(param, "tag")
		if tag == "" {
			tag = "v1.0.0"
		}

		// 获取客户端显式传入的 IP
		clientIP := utils.GetQueryParamCaseInsensitive(param, "client_ip")
		if clientIP == "" {
			http.Error(w, "Client IP is required (please pass ?client_ip=xxx)", http.StatusBadRequest)
			return
		}

		//// 订阅记录的客户端地址
		////clientAddr := r.RemoteAddr
		//// 提取 IP 地址，去掉端口
		//host, _, err := net.SplitHostPort(r.RemoteAddr)
		//if err != nil {
		//	http.Error(w, "Error parsing remote address", http.StatusInternalServerError)
		//	return
		//}
		//// 如果 host 是 IPv6 地址，需要加上方括号
		//if strings.Contains(host, ":") {
		//	host = "[" + host + "]"
		//}
		clientAddr := fmt.Sprintf("http://%s:8080/receive?filename=%s", clientIP, fileName)

		// 调用订阅管理器记录订阅信息
		d.Subscribers.Subscribe(fileName, tag, clientAddr)

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Subscription successful"))

	}
}
