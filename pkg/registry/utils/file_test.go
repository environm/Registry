package utils

import (
	"bytes"
	"os"
	"testing"
)

// 帮助函数：测试结束后清理生成的文件
func cleanup(dest, fileName, tag string) {
	filePath := generateFilePath(dest, fileName, tag)
	os.Remove(filePath)
}

func TestSaveFile(t *testing.T) {
	// 测试保存文件功能
	dest := "./test_dir"                   // 目标目录
	fileName := "testFile"                 // 文件名
	tag := "v1"                            // 标签
	data := []byte("This is a test file.") // 文件内容
	file := NewFile(fileName, tag, false)  // 创建文件对象

	// 保存文件
	err := file.SaveFile(dest, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// 检查文件是否被正确创建
	expectedPath := generateFilePath(dest, fileName, tag)
	_, err = os.Stat(expectedPath)
	if os.IsNotExist(err) {
		t.Fatalf("Expected file to be created, but it does not exist")
	}

	// 验证文件大小是否正确
	if file.Size != int64(len(data)) {
		t.Fatalf("Expected file size to be %d, but got %d", len(data), file.Size)
	}

	// 清理测试环境
	//cleanup(dest, fileName, tag)
}

func TestLoadFile(t *testing.T) {
	// 测试加载文件功能
	dest := "./test_dir"
	fileName := "testFile"
	tag := "v1"
	data := []byte("This is a test file.") // 文件内容
	file := NewFile(fileName, tag, false)

	// 先保存文件
	err := file.SaveFile(dest, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// 测试加载文件
	loadedFile, err := file.LoadFile(dest)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	defer loadedFile.Close() // 确保文件正确关闭

	// 清理测试环境
	cleanup(dest, fileName, tag)
}

func TestDeleteFile(t *testing.T) {
	// 测试删除文件功能
	dest := "./test_dir"
	fileName := "testFile"
	tag := "v1"
	data := []byte("This is a test file.")
	file := NewFile(fileName, tag, false)

	// 先保存文件
	err := file.SaveFile(dest, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// 测试删除文件
	err = file.DeleteFile(dest)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// 检查文件是否已删除
	expectedPath := generateFilePath(dest, fileName, tag)
	_, err = os.Stat(expectedPath)
	if !os.IsNotExist(err) {
		t.Fatalf("Expected file to be deleted, but it still exists")
	}
}

func TestUpdateFile(t *testing.T) {
	// 测试更新文件功能
	dest := "./test_dir"
	fileName := "testFile"
	tag := "v1"
	data := []byte("This is the original file.")       // 原始内容
	updatedData := []byte("This is the updated file.") // 更新后的内容
	file := NewFile(fileName, tag, false)

	// 先保存文件
	err := file.SaveFile(dest, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// 更新文件内容
	err = file.UpdateFile(dest, bytes.NewReader(updatedData))
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// 加载更新后的文件
	loadedFile, err := file.LoadFile(dest)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	defer loadedFile.Close()

	// 清理测试环境
	cleanup(dest, fileName, tag)
}
