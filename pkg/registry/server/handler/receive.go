package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
	"path/filepath"
)

// ReceiveHandler 文件/文件夹接收
// 方法 Post
// URL /receive

type ReceiveHandler struct {
	//
	DataPath string
	//
	FileMapping *data.FileMapping
	//
	Subscribers *data.SubscriptionManager
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *ReceiveHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewReceiveHandler(dataPath string, fileMapping *data.FileMapping, subscribers *data.SubscriptionManager) *ReceiveHandler {
	dh := &ReceiveHandler{
		DataPath:    dataPath,
		FileMapping: fileMapping,
		Subscribers: subscribers,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &ReceiveHandler{}

func (d *ReceiveHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// 四种情况：文件夹订阅、文件订阅、文件夹上传、文件上传
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
			return
		}
		// 获取参数
		param := r.URL.Query()
		fileName := utils.GetQueryParamCaseInsensitive(param, "fileName")
		if fileName == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			return
		}
		fileName = filepath.Clean(fileName)
		// 获取标签（如果没有提供，则设置默认标签）
		tag := utils.GetQueryParamCaseInsensitive(param, "tag")
		if tag == "" {
			tag = "v1.0.0" // 默认标签
		}

		fileType := r.Header.Get("FileType")
		if fileType == "" {
			http.Error(w, "FileType header is required", http.StatusBadRequest)
			return
		}

		// 处理订阅
		if d.Subscribers.IsSubscribed(fileName, tag) {
			// 获取订阅者
			subscribers := d.Subscribers.GetSubscribers(fileName, tag)
			for _, subscriber := range subscribers {
				// 转发给订阅者
				if err := utils.ForwardRequest(r, subscriber, w, fileName, tag); err != nil {
					http.Error(w, fmt.Sprintf("Failed to forward request: %v", err), http.StatusInternalServerError)
					return
				}
			}
			return
		}

		// 如果没有订阅请求，则存储到文件系统中，等待请求下载
		switch fileType {
		case "folder":
			utils.ReceiveDir(w, r, d.DataPath)
			err := d.FileMapping.SaveFolder(d.DataPath, fileName, tag, false)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to save folder: %v", err), http.StatusInternalServerError)
				return
			}
		case "file":
			err := d.FileMapping.SaveFile(d.DataPath, fileName, tag, false, r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("File received successfully"))
	}
}
