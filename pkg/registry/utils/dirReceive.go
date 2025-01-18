package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// createDir 创建目录
func createDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dirPath, err)
		}
		fmt.Printf("Directory created: %s\n", dirPath)
	} else {
		fmt.Printf("Directory already exists: %s\n", dirPath)
	}
	return nil
}

//func buildPath(baseDir, parentPath, name string) string {
//	switch {
//	case rootPath == parentPath && parentPath == name:
//		return filepath.Join(baseDir, rootPath) // 所有相等，返回一个
//	case rootPath == parentPath:
//		return filepath.Join(baseDir, rootPath, name) // rootPath 和 parentPath 相等，合并 rootPath 和 name
//	default:
//		return filepath.Join(baseDir, rootPath, parentPath, name) // 都不相等，返回完整路径
//	}
//}

// ReceiveDir 接收文件和目录信息，并正确创建或存储
func ReceiveDir(w http.ResponseWriter, r *http.Request, baseDir string) {
	// 确保 BaseDir 不为空
	if baseDir == "" {
		http.Error(w, "BaseDir is not set", http.StatusInternalServerError)
		return
	}

	// 确保 BaseDir 路径存在
	baseDir = filepath.Clean(baseDir)
	err := createDir(baseDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create base directory: %v", err), http.StatusInternalServerError)
		return
	}

	contentType := r.Header.Get("Content-Type")
	rootPath := r.Header.Get("rootPath") // 根目录路径
	parentPath := r.Header.Get("Parent-Path")

	if rootPath == "" || parentPath == "" {
		http.Error(w, "Root path or parent path is missing", http.StatusBadRequest)
		return
	}

	rootPath = filepath.Clean(rootPath)
	parentPath = filepath.Clean(parentPath)

	if contentType == "application/text" {
		// 处理目录
		dirName := r.Header.Get("Directory-Name")
		if dirName == "" {
			http.Error(w, "Directory name is missing", http.StatusBadRequest)
			return
		}
		// 构建目标路径
		dirPath := filepath.Join(baseDir, parentPath, dirName)

		err := createDir(dirPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create directory: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Directory created successfully"))
	} else if contentType == "application/octet-stream" {
		// 处理文件
		fileName := r.Header.Get("File-Name")
		if fileName == "" {
			http.Error(w, "File name is missing", http.StatusBadRequest)
			return
		}

		// 构建目标路径
		filePath := filepath.Join(baseDir, parentPath, fileName)
		parentFullPath := filepath.Dir(filePath)

		err := createDir(parentFullPath) // 确保父目录存在
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create parent directory: %v", err), http.StatusInternalServerError)
			return
		}

		// 创建并写入文件
		file, err := os.Create(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create file: %v", err), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		_, err = io.Copy(file, r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to write file: %v", err), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("File uploaded successfully"))
	} else {
		http.Error(w, "Unsupported content type", http.StatusBadRequest)
	}
}
