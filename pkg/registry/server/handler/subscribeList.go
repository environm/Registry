package handler

import (
	"encoding/json"
	"hit.edu/framework/pkg/registry/data"
	"net/http"
)

type SubscribeListHandler struct {
	//
	Subscribers *data.SubscriptionManager
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *SubscribeListHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewScribeListHandler(subscribers *data.SubscriptionManager) *SubscribeListHandler {
	dh := &SubscribeListHandler{
		Subscribers: subscribers,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &SubscribeHandler{}

func (d *SubscribeListHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is supported", http.StatusMethodNotAllowed)
			return
		}
		//lists := d.FileMapping.ListFiles()
		lists := d.Subscribers.GetSubscriptionList()

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
