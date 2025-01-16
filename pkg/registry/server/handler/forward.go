package handler

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/utils"
	"io"
	"net/http"
	"path/filepath"
)

// HandlePostAndForward 处理 POST 请求并实时转发数据
// TODO:

type ForwardHandler struct {
	//
	DataPath string
	//
	FileMapping *utils.FileMapping
	//
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *ForwardHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewForwardHandler(dataPath string, fileMapping *utils.FileMapping) *ForwardHandler {
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
		//fileName := r.URL.Query().Get("filename")
		fileName := utils.GetQueryParamCaseInsensitive(param, "filename")
		if fileName == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			logs.Infof("Filename missing in request")
			return
		}
		fileName = filepath.Clean(fileName)

		// 获取标签
		//tag := r.URL.Query().Get("tag")
		tag := utils.GetQueryParamCaseInsensitive(param, "tag")
		if tag == "" {
			tag = "v1.0.0" // 设置默认标签
		}

		// 获取目标地址
		//targetURL := r.URL.Query().Get("target")
		targetURL := utils.GetQueryParamCaseInsensitive(param, "target")
		if targetURL == "" {
			http.Error(w, "Target URL is required for forwarding", http.StatusBadRequest)
			logs.Infof("Target URL missing in request")
			return
		}

		// 在目标 URL 中添加 filename 和 tag 参数
		targetURL = fmt.Sprintf("%s/post?filename=%s&tag=%s", targetURL, fileName, tag)
		logs.Infof("Forwarding to URL: %s", targetURL)

		// 创建 File 实例并加载文件
		//f := utils.NewFile(fileName, tag)
		file, err := d.FileMapping.LoadFile(fileName, tag, d.DataPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to load file: %v", err), http.StatusNotFound)
			logs.Infof("Error loading file %s: %v", fileName, err)
			return
		}
		defer file.Close()

		// 重置文件指针，确保从文件头开始读取
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			http.Error(w, "Failed to prepare file for reading", http.StatusInternalServerError)
			logs.Infof("Failed to reset file pointer for %s: %v", fileName, err)
			return
		}

		// 创建 HTTP 请求
		req, err := http.NewRequest(http.MethodPost, targetURL, file)
		if err != nil {
			http.Error(w, "Failed to create HTTP request", http.StatusInternalServerError)
			logs.Infof("Failed to create HTTP POST request to %s: %v", targetURL, err)
			return
		}

		// 设置 Content-Type
		req.Header.Set("Content-Type", "multipart/form-data")

		// 发送请求
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to forward data to target", http.StatusInternalServerError)
			logs.Infof("Failed to forward data to %s: %v", targetURL, err)
			return
		}
		defer resp.Body.Close()

		// 记录转发结果
		logs.Infof("Data successfully forwarded to %s with status %d", targetURL, resp.StatusCode)

		// 检查目标服务器的响应状态码
		if resp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("Target server responded with status %d", resp.StatusCode), http.StatusBadGateway)
			logs.Infof("Target server responded with status %d for file %s", resp.StatusCode, fileName)
			return
		}

		// 返回目标服务器的响应内容
		w.WriteHeader(http.StatusOK)
		io.Copy(w, resp.Body)
		logs.Infof("Data transfer completed for file %s", fileName)

	}
}
