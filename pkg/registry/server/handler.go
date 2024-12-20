package server

import (
	"hit.edu/framework/pkg/registry/server/handler"
	"net/http"
	"time"
)

const (
	defaultKeepAlivePeriod = 3 * time.Minute
)

// 处理不同的HTTP请求
type RegistryHandler struct {
	// 处理文件上传
	UploadHandler handler.Handler
	// 处理文件下载
	DownloadHandler handler.Handler
	// 处理文件转发
	ForwardHandler handler.Handler
	// 处理文件删除
	// TODO:
	DeleteHandler handler.Handler
	// 处理文件查询
	// TODO:
	QueryIsExistsHandler handler.Handler
}

func NewRegistryHandler(dataPath string) *RegistryHandler {
	// TODO: Download等改成Handler, 实现ServeHTTP等函数
	rh := &RegistryHandler{
		UploadHandler:        handler.NewUploadHandler(dataPath),
		DownloadHandler:      handler.NewDownloadHandler(dataPath),
		ForwardHandler:       handler.NewForwardHandler(dataPath),
		DeleteHandler:        handler.NewDeleteHandler(dataPath),
		QueryIsExistsHandler: handler.NewQueryIsExistsHandler(dataPath),
	}

	// TODO: 临时用法,注册路由
	http.HandleFunc("/download", rh.DownloadHandler.GetHandler())
	http.HandleFunc("/upload", rh.UploadHandler.GetHandler())
	http.HandleFunc("/forward", rh.ForwardHandler.GetHandler())
	http.HandleFunc("/delete", rh.DeleteHandler.GetHandler())
	http.HandleFunc("/query/exits", rh.QueryIsExistsHandler.GetHandler())

	return rh
}
