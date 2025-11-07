package handler

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"io"
	"net"
	"net/http"
)

// DownloadHandler 对应文件下载请求
// 方法 GET
// URL /download
// Param filename
// TODO:
type DownloadHandler struct {
	//
	DataPath string
	//
	//FileMapping *data.FileMapping
	DataSpecList *data.DataSpecList
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *DownloadHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewDownloadHandler(dataPath string, dataSpecList *data.DataSpecList) *DownloadHandler {
	dh := &DownloadHandler{
		DataPath:     dataPath,
		DataSpecList: dataSpecList,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &DownloadHandler{}

func (d *DownloadHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is supported", http.StatusMethodNotAllowed)
			return
		}

		fileName, tag, _, fileType, err := utils.GetFileParams(r, "v1.0.0")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting file parameters: %v", err), http.StatusBadRequest)
			return
		}

		//clientAddr := r.RemoteAddr
		// 提取 IP 地址，去掉端口
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Error parsing remote address", http.StatusInternalServerError)
			return
		}
		clientURL := fmt.Sprintf("http://%s:8080/receive?filename=%s", host, fileName)

		switch fileType {
		case "folder":
			folderPath := d.DataSpecList.GetFilePath(fileName, tag)
			//utils.Traverse("D:\\Programming\\GolandProjects\\Registry\\tmp\\data\\downloads", "http://localhost:8080/receive?filename=downloads")

			err := utils.Traverse(folderPath, clientURL)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error traversing directory: %v", err), http.StatusInternalServerError)
				return
			}
		case "file":
			//file, err := d.FileMapping.LoadFile(fileName, tag, d.DataPath)
			file, err := d.DataSpecList.LoadFile(fileName, tag)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error loading file: %v", err), http.StatusNotFound)
				return
			}
			defer file.Close()
			// 设置响应头，支持文件下载
			//w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			//w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(fileName)))
			w.WriteHeader(http.StatusOK)
			// 将文件内容写入响应
			_, err = io.Copy(w, file)
			if err != nil {
				http.Error(w, "Failed to send file", http.StatusInternalServerError)
				return
			}
			logs.Infof("File %s (tag: %s) downloaded successfully", fileName, tag)

		default:
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		// 检查文件是否标记为永久存储
		//isExits, _ := d.FileMapping.QueryFile(fileName, tag)
		isExits, _ := d.DataSpecList.GetDataSpec(fileName, tag)
		fmt.Printf("IsPermanent: %v", isExits)
		if !isExits.IsPermanent {
			// 文件未标记为永久存储，下载后删除文件
			err := d.DataSpecList.DeleteFile(fileName, tag)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete file after download: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}
}
