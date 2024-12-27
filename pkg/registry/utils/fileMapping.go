package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

type FileMapping struct {
	files map[string]File // 文件存储器，key 为 文件名 + tag，value 为文件信息
	mutex sync.Mutex
}

func NewFileMapping() *FileMapping {
	return &FileMapping{
		files: make(map[string]File),
	}
}

// 转化为 josn 文件
func (fm *FileMapping) SaveToFile(filePath string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	data, err := json.MarshalIndent(fm.files, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
	//return os.WriteFile(filePath, data, 0644)
}

// 加载 josn 文件至 内存
func (fm *FileMapping) LoadFromFile(filePath string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在时，初始化一个空的文件映射
			fm.files = make(map[string]File)
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, &fm.files); err != nil {
		return fmt.Errorf("failed to unmarshal file mapping: %v", err)
	}

	return nil
}

// getFileKey 获取文件 key
func (fm *FileMapping) getFileKey(fileName, tag string) string {
	return fmt.Sprintf("%s-%s", fileName, tag)
}

// SaveFile 添加保存新文件
func (fm *FileMapping) SaveFile(dest, fileName, tag string, isPermanent bool, data io.Reader) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	key := fm.getFileKey(fileName, tag)
	if _, exists := fm.files[key]; exists {
		return fmt.Errorf("file %s already exists", key)
	}

	// 创建文件对象
	file := NewFile(fileName, tag, isPermanent)
	err := file.SaveFile(dest, data)
	if err != nil {
		return err
	}

	// 将文件信息添加到映射中
	fm.files[key] = *file
	return nil
}

// LoadFile 根据文件名和标签加载文件
func (fm *FileMapping) LoadFile(fileName, tag string, dest string) (*os.File, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	key := fm.getFileKey(fileName, tag)
	file, exists := fm.files[key]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	// 加载文件
	return file.LoadFile(dest)
}

// DeleteFile 根据文件名和标签删除文件
func (fm *FileMapping) DeleteFile(fileName, tag string, dest string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	key := fm.getFileKey(fileName, tag)
	file, exists := fm.files[key]
	if !exists {
		return fmt.Errorf("file not found: %s", key)
	}

	// 删除文件
	err := file.DeleteFile(dest)
	if err != nil {
		return err
	}

	// 从映射中删除文件信息
	delete(fm.files, key)
	return nil
}

// UpdateFile 根据文件名和标签更新文件信息
func (fm *FileMapping) UpdateFile(fileName, tag string, data io.Reader, dest string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	key := fm.getFileKey(fileName, tag)
	file, exists := fm.files[key]
	if !exists {
		return fmt.Errorf("file not found: %s", key)
	}

	// 更新文件
	err := file.UpdateFile(dest, data)
	if err != nil {
		return err
	}

	fm.files[key] = file
	return nil
}

// QueryFile 根据文件名和标签查询文件
func (fm *FileMapping) QueryFile(fileName, tag string) (*File, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	key := fm.getFileKey(fileName, tag)
	file, exists := fm.files[key]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	return &file, nil
}

// ListFiles 获取所有文件的列表
func (fm *FileMapping) ListFiles() []File {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	var filesList []File
	for _, file := range fm.files {
		filesList = append(filesList, file)
	}
	return filesList
}
