package handler

import (
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
	Subscribers *utils.SubscriptionManager
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *SubscribeHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewSubscribeHandler(subscribers *utils.SubscriptionManager) *SubscribeHandler {
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

		// 订阅记录的客户端地址
		clientAddr := r.RemoteAddr

		// 调用订阅管理器记录订阅信息
		d.Subscribers.Subscribe(fileName, tag, clientAddr)

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Subscription successful"))

	}
}
