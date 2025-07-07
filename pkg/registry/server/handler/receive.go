package handler

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
)

// ReceiveHandler 文件/文件夹接收
// 方法 Post
// URL /receive

type ReceiveHandler struct {
	//
	DataPath string
	//
	//FileMapping *data.FileMapping
	DataSpecList *data.DataSpecList
	//
	Subscribers *data.SubscriptionManager
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *ReceiveHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

// NewReceiveHandler 创建一个 ReceiveHandler 实例
func NewReceiveHandler(dataPath string, dataSpecList *data.DataSpecList, subscribers *data.SubscriptionManager) *ReceiveHandler {
	dh := &ReceiveHandler{
		DataPath:     dataPath,
		DataSpecList: dataSpecList,
		Subscribers:  subscribers,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &ReceiveHandler{}

func (d *ReceiveHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// 判断流量类型
		flowType := r.Header.Get("FlowType")
		// etcd 的流量
		if flowType == "etcd" {
			targetURL, err := utils.TransformTargetURL(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			logs.Infof("Forwarding to URL: %s", targetURL)
			// 转发
			if err := utils.ForwardRequest(r, targetURL, w); err != nil {
				http.Error(w, fmt.Sprintf("Failed to forward request: %v", err), http.StatusInternalServerError)
				return
			}
		} else {
			// 四种情况：文件夹订阅、文件订阅、文件夹上传、文件上传
			fileName, tag, owner, fileType, err := utils.GetFileParams(r, "v1.0.0")
			if err != nil {
				http.Error(w, fmt.Sprintf("Error getting file parameters: %v", err), http.StatusBadRequest)
				return
			}

			// 处理订阅
			if d.Subscribers.IsSubscribed(fileName, tag) {
				// 获取订阅者
				subscribers := d.Subscribers.GetSubscribers(fileName, tag)
				for _, subscriber := range subscribers {
					// 转发给订阅者
					if err := utils.ForwardRequest(r, subscriber, w); err != nil {
						http.Error(w, fmt.Sprintf("Failed to forward request: %v", err), http.StatusInternalServerError)
						return
					}
				}
				return
			}

			// 如果没有订阅请求，则存储到文件系统中，等待请求下载
			switch fileType {
			case "folder", "completion":
				utils.ReceiveDir(w, r, d.DataPath)
				//err := d.DataSpecList.SaveFolder(d.DataPath, fileName, tag, false)
				_, err := d.DataSpecList.SaveFolder(d.DataPath, fileName, tag, owner, false)
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to save folder: %v", err), http.StatusInternalServerError)
					return
				}
			case "file":
				//err := d.DataSpecList.SaveFile(d.DataPath, fileName, tag, false, r.Body)
				_, err := d.DataSpecList.SaveFile(d.DataPath, fileName, tag, owner, false, r.Body)
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
}
