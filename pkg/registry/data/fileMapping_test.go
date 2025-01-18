package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestAddFile(t *testing.T) {
	fm := NewFileMapping()
	dest := "./test_dir"
	fileName := "test.txt"
	tag := "v1"
	content := "Hello, World!"
	err := fm.SaveFile(dest, fileName, tag, false, bytes.NewReader([]byte(content)))
	if err != nil {
		t.Errorf("SaveFile failed: %v", err)
	}

	file, err := fm.QueryFile(fileName, tag)
	if err != nil {
		t.Errorf("QueryFile failed: %v", err)
	}
	if file == nil || file.FileName != fileName || file.Tag != tag {
		t.Errorf("QueryFile returned incorrect file: %+v", file)
	}
}

func TestUpdateFile(t *testing.T) {
	fm := NewFileMapping()
	dest := "./test_dir"
	fileName := "test.txt"
	tag := "v1"
	content := "Hello, World!"
	_ = fm.SaveFile(dest, fileName, tag, false, bytes.NewReader([]byte(content)))

	newContent := "Updated Content!"
	err := fm.UpdateFile(fileName, tag, bytes.NewReader([]byte(newContent)), dest)
	if err != nil {
		t.Errorf("UpdateFile failed: %v", err)
	}

	updatedFile, err := fm.QueryFile(fileName, tag)
	if err != nil {
		t.Errorf("QueryFile after update failed: %v", err)
	}
	fileData, err := os.ReadFile(updatedFile.FilePath)
	if err != nil {
		t.Errorf("failed to read updated file: %v", err)
	}
	if string(fileData) != newContent {
		t.Errorf("file content mismatch: got %s, want %s", fileData, newContent)
	}
}

func TestListFiles(t *testing.T) {
	fm := NewFileMapping()
	dest := "./test_dir"
	fileName := "test.txt"
	tag := "v1"
	content := "Hello, World!"
	_ = fm.SaveFile(dest, fileName, tag, false, bytes.NewReader([]byte(content)))
	_ = fm.SaveFile(dest, fileName, "v2", false, bytes.NewReader([]byte(content)))

	files := fm.ListFiles()
	// 输出文件列表
	for _, file := range files {
		fileData, _ := json.MarshalIndent(file, "", "  ")
		fmt.Printf("File: %s\n", string(fileData))
	}
	//if len(files) != 1 || files[0].FileName != fileName {
	//	t.Errorf("ListFiles returned incorrect files: %+v", files)
	//}
}

func TestDeleteFile(t *testing.T) {
	fm := NewFileMapping()
	dest := "./test_dir"
	fileName := "test.txt"
	tag := "v1"
	content := "Hello, World!"
	_ = fm.SaveFile(dest, fileName, tag, false, bytes.NewReader([]byte(content)))

	err := fm.DeleteFile(fileName, tag, dest)
	if err != nil {
		t.Errorf("DeleteFile failed: %v", err)
	}

	_, err = fm.QueryFile(fileName, tag)
	if err == nil {
		t.Errorf("QueryFile should fail for deleted file")
	}
}

func TestSaveFileMapping(t *testing.T) {
	fm := NewFileMapping()
	dest := "./test_dir"
	fm.SaveFile(dest, "test.txt", "v1", false, bytes.NewReader([]byte("Hello, World!")))
	fm.SaveFile(dest, "test2.txt", "v2", false, bytes.NewReader([]byte("Hello, World2!")))
	jsonPath := dest + "/file_mapping.json"
	if err := fm.SaveToFile(jsonPath); err != nil {
		t.Errorf("SaveToFile failed: %v", err)
	}

	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Errorf("file mapping JSON file was not created")
	}
}

func TestLoadFileMapping(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_file_mapping")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 准备测试数据并保存为 JSON 文件
	fm := NewFileMapping()
	fm.SaveFile(tempDir, "test1.txt", "v1", false, bytes.NewReader([]byte("Content1")))
	fm.SaveFile(tempDir, "test2.txt", "v2", false, bytes.NewReader([]byte("Content2")))
	jsonPath := tempDir + "/file_mapping.json"
	if err := fm.SaveToFile(jsonPath); err != nil {
		t.Fatalf("failed to save file mapping: %v", err)
	}

	// 加载 JSON 文件并打印内容
	loadedFM := NewFileMapping()
	if err := loadedFM.LoadFromFile(jsonPath); err != nil {
		t.Errorf("LoadFromFile failed: %v", err)
	}

	// 打印加载的文件映射
	loadedFiles := loadedFM.ListFiles()
	for _, file := range loadedFiles {
		fileData, _ := json.MarshalIndent(file, "", "  ")
		fmt.Printf("Loaded File: %s\n", string(fileData))
	}
}
