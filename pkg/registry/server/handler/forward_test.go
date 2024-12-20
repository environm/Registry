package handler

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestLoadFile(t *testing.T) {
	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		// 解析 Query 参数
		fileName := r.URL.Query().Get("filename")
		tag := r.URL.Query().Get("tag")

		// 检查 filename 和 tag 是否存在
		if fileName == "" || tag == "" {
			http.Error(w, "Missing filename or tag", http.StatusBadRequest)
			return
		}

		// 读取请求体内容
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		// 打印收到的内容
		fmt.Printf("Received file: %s (tag: %s)\n", fileName, tag)
		fmt.Printf("Content: %s\n", string(body))

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("File successfully received"))
	})

	fmt.Println("Starting mock target server on port 8082...")
	http.ListenAndServe(":8082", nil)
}
