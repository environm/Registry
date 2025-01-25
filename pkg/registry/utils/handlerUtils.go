package utils

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// GetQueryParamCaseInsensitive 获取不区分大小写的查询参数
func GetQueryParamCaseInsensitive(params map[string][]string, paramName string) string {
	for key, values := range params {
		if strings.EqualFold(key, paramName) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

// GetFileParams 从请求中提取文件名和标签参数
func GetFileParams(r *http.Request, defaultTag string) (string, string, string, string, error) {
	// 获取参数
	param := r.URL.Query()

	// 获取文件名参数
	fileName := GetQueryParamCaseInsensitive(param, "filename")
	if fileName == "" {
		return "", "", "", "", fmt.Errorf("filename is required")
	}
	// 清理路径，防止路径遍历攻击
	fileName = filepath.Clean(fileName)

	// 获取标签参数（如果未提供，使用默认值）
	tag := GetQueryParamCaseInsensitive(param, "tag")
	if tag == "" {
		tag = defaultTag
	}

	// 获取文件所有者
	owner := r.RemoteAddr

	// 获取文件类型
	fileType := r.Header.Get("FileType")
	if fileType == "" {
		fileType = "file"
	}

	return fileName, tag, owner, fileType, nil
}

// ForwardRequest 将 HTTP 请求转发给订阅者
func ForwardRequest(r *http.Request, url string, w http.ResponseWriter, fileName, tag string) error {
	// 创建一个新的请求，将原始请求内容复制到新请求中
	req, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		http.Error(w, "Failed to create HTTP request", http.StatusInternalServerError)
		logs.Infof("Failed to create HTTP POST request to %s: %v", url, err)
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 复制原始请求的头部到新请求中
	req.Header = r.Header.Clone()

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to forward data to target", http.StatusInternalServerError)
		logs.Infof("Failed to forward data to %s: %v", url, err)
		return fmt.Errorf("failed to forward data: %w", err)
	}
	defer resp.Body.Close()

	// 检查目标服务器的响应状态码
	if resp.StatusCode != http.StatusCreated {
		http.Error(w, fmt.Sprintf("Target server responded with status %d", resp.StatusCode), http.StatusBadGateway)
		logs.Infof("Target server responded with status %d for file %s_%s", resp.StatusCode, fileName, tag)
		return fmt.Errorf("target server responded with status %d", resp.StatusCode)
	}

	// 设置状态码并将响应体写回客户端
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logs.Infof("Error writing response: %v", err)
		return fmt.Errorf("error writing response: %w", err)
	}

	// 记录成功转发的日志
	logs.Infof("Data successfully forwarded to %s with status %d", url, resp.StatusCode)
	return nil
}
