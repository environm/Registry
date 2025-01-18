package server

import (
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/server/handler"
	"net/http"
	"time"
)

const (
	defaultKeepAlivePeriod = 3 * time.Minute
)

// RegistryHandler 处理不同的HTTP请求
type RegistryHandler struct {
	// 处理文件上传
	UploadHandler handler.Handler
	// 处理文件下载
	DownloadHandler handler.Handler
	// 处理文件转发
	ForwardHandler handler.Handler
	// 处理文件接收
	ReceiveHandler handler.Handler
	// 处理文件删除
	// TODO:
	DeleteHandler handler.Handler
	// 处理文件查询
	// TODO:
	QueryIsExistsHandler handler.Handler
	// 处理文件列表查询
	QueryListHandler handler.Handler
	// 处理文件订阅
	SubscribeHandler handler.Handler
	// 处理文件目录上传
	CatalogueUploadHandler handler.Handler
	// 处理文件目录下载
	CatalogueDownloadHandler handler.Handler
}

func NewRegistryHandler(dataPath string, fileMapping *data.FileMapping, subscribers *data.SubscriptionManager) *RegistryHandler {
	// TODO: Download等改成Handler, 实现ServeHTTP等函数
	rh := &RegistryHandler{
		UploadHandler:            handler.NewUploadHandler(dataPath, fileMapping),
		DownloadHandler:          handler.NewDownloadHandler(dataPath, fileMapping),
		ForwardHandler:           handler.NewForwardHandler(dataPath, fileMapping),
		ReceiveHandler:           handler.NewReceiveHandler(dataPath, fileMapping, subscribers),
		DeleteHandler:            handler.NewDeleteHandler(dataPath, fileMapping),
		QueryIsExistsHandler:     handler.NewQueryIsExistsHandler(dataPath, fileMapping),
		QueryListHandler:         handler.NewQueryListHandler(dataPath, fileMapping),
		SubscribeHandler:         handler.NewSubscribeHandler(subscribers),
		CatalogueUploadHandler:   handler.NewCatalogueUploadHandler(dataPath, fileMapping),
		CatalogueDownloadHandler: handler.NewCatalogueDownloadHandler(dataPath, fileMapping),
	}

	// TODO: 临时用法,注册路由
	http.HandleFunc("/download", rh.DownloadHandler.GetHandler())
	http.HandleFunc("/upload", rh.UploadHandler.GetHandler())
	http.HandleFunc("/forward", rh.ForwardHandler.GetHandler())
	http.HandleFunc("/receive", rh.ReceiveHandler.GetHandler())
	http.HandleFunc("/delete", rh.DeleteHandler.GetHandler())
	http.HandleFunc("/query/exits", rh.QueryIsExistsHandler.GetHandler())
	http.HandleFunc("/query/list", rh.QueryListHandler.GetHandler())
	http.HandleFunc("/subscribe", rh.SubscribeHandler.GetHandler())
	http.HandleFunc("/catalogueUpload", rh.CatalogueUploadHandler.GetHandler())
	http.HandleFunc("/catalogueDownload", rh.CatalogueDownloadHandler.GetHandler())
	return rh
}
