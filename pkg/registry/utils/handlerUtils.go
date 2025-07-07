package utils

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
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
func ForwardRequest(r *http.Request, newURL string, w http.ResponseWriter) error {
	logs.Infof("Forwarding to subscriber: %s", newURL)

	// 创建一个新的请求，将原始请求内容复制到新请求中
	req, err := http.NewRequest(r.Method, newURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create HTTP request", http.StatusInternalServerError)
		logs.Infof("Failed to create HTTP POST request to %s: %v", newURL, err)
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 复制原始请求的头部到新请求中
	req.Header = r.Header.Clone()

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to forward data to target", http.StatusInternalServerError)
		logs.Infof("Failed to forward data to %s: %v", newURL, err)
		return fmt.Errorf("failed to forward data: %w", err)
	}
	defer resp.Body.Close()

	// 检查目标服务器的响应状态码
	if resp.StatusCode >= 400 {
		http.Error(w, fmt.Sprintf("Target server responded with status %d", resp.StatusCode), http.StatusBadGateway)
		logs.Infof("Target server responded with status %d", resp.StatusCode)
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
	logs.Infof("Data successfully forwarded to %s with status %d", newURL, resp.StatusCode)
	return nil
}

// TransformURL 将请求的 URL 转换为接收数据的 URL
func TransformURL(ClusterID string, r *http.Request) (string, error) {
	// 解析原始 URL
	parsedURL, err := url.Parse(r.URL.String())
	if err != nil {
		logs.Infof("Failed to parse request URL: %v", err)
		return "", fmt.Errorf("failed to parse request URL")
	}

	// 解析 Host 替换 ClusterID
	originalHost, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		originalHost = r.Host
		port = ""
	}
	hostParts := strings.SplitN(originalHost, ".", 2)
	if len(hostParts) < 2 {
		logs.Infof("Invalid host format: %s", originalHost)
		return "", fmt.Errorf("invalid host format")
	}

	// 替换 cluster ID
	hostParts[0] = ClusterID
	newHost := strings.Join(hostParts, ".")
	if port != "" {
		newHost = net.JoinHostPort(newHost, port)
	}

	// 替换路径 `/forward` 为 `/receive`
	parsedURL.Path = strings.Replace(parsedURL.Path, "/forward", "/receive", 1)

	// 构造新的完整 URL
	newURL := url.URL{
		Scheme:   "http",
		Host:     newHost,
		Path:     parsedURL.Path,
		RawQuery: parsedURL.Query().Encode(),
	}

	return newURL.String(), nil
}

// TransformTargetURL 将请求中的 target 参数解析并转换为目标 URL
func TransformTargetURL(r *http.Request) (string, error) {
	// 解析请求 URL
	parsedURL, err := url.Parse(r.URL.String())
	if err != nil {
		log.Printf("Failed to parse request URL: %v", err)
		return "", fmt.Errorf("failed to parse request URL")
	}

	// 获取查询参数中的 target 字段
	targetURLStr := parsedURL.Query().Get("target")
	if targetURLStr == "" {
		log.Println("Target URL missing in the request")
		return "", fmt.Errorf("target URL is missing in the request")
	}

	// 解析 target URL
	targetURL, err := url.Parse(targetURLStr)
	if err != nil {
		log.Printf("Failed to parse target URL: %v", err)
		return "", fmt.Errorf("failed to parse target URL")
	}

	// 返回目标 URL 字符串
	return targetURL.String(), nil
}
