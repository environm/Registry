package data

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileType string

const (
	FileTypeFile   FileType = "file"
	FileTypeFolder FileType = "folder"
)

// DataSpec 文件元数据结构
type DataSpec struct {
	FileName    string    `json:"file_name"`    // 文件名称
	Tag         string    `json:"tag"`          // 文件版本号
	Type        FileType  `json:"type"`         // 文件类型
	IsPermanent bool      `json:"is_permanent"` // 是否持久化存储
	FilePath    string    `json:"file_path"`    // 文件存储路径
	StorageTime time.Time `json:"storage_time"` // 存储时间
	Owner       string    `json:"owner"`        // 文件所有者
	Size        int64     `json:"size"`         // 文件大小
	FileHash    string    `json:"file_hash"`    // 文件哈希值
}

type DataSpecList struct {
	DataSpecIndex map[string]DataSpec
	mutex         sync.Mutex
}

func NewDataSpecList() *DataSpecList {
	return &DataSpecList{
		DataSpecIndex: make(map[string]DataSpec),
	}
}

func NewDataSpec(fileName, tag, owner string, fileType FileType, isPermanent bool) *DataSpec {
	return &DataSpec{
		FileName:    fileName,
		Tag:         tag,
		Type:        fileType,
		FilePath:    "",
		IsPermanent: isPermanent,
		StorageTime: time.Now(),
		FileHash:    "",
		Owner:       owner,
		Size:        0,
	}
}

// generateFilePath 根据文件路径、文件名和标签生成文件路径
func generateFilePath(dest, fileName, tag string) string {
	return filepath.Join(dest, fmt.Sprintf("%s_%s", fileName, tag))
}

// getKey 获取文件 key
func getKey(fileName, tag string) string {
	return fmt.Sprintf("%s_%s", fileName, tag)
}

// fileExists 检查文件是否存在
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// calculateFileHash 计算文件的 SHA-256 哈希值
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}

// SaveFile 保存文件到指定目录，并更新 DataSpecIndex
func (dsl *DataSpecList) SaveFile(dest, fileName, tag, owner string, isPermanent bool, data io.Reader) (*DataSpec, error) {
	dsl.mutex.Lock()
	defer dsl.mutex.Unlock()
	// 生成文件路径
	filePath := generateFilePath(dest, fileName, tag)
	// 检查文件是否存在
	if fileExists(filePath) {
		return nil, fmt.Errorf("file already exists: %s_%s", fileName, tag)
	}
	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	written, err := io.Copy(file, data)
	if err != nil {
		return nil, err
	}
	// 计算文件哈希值
	fileHash, err := calculateFileHash(filePath)
	if err != nil {
		return nil, err
	}
	// 创建 DataSpec
	d := DataSpec{
		FileName:    fileName,
		Tag:         tag,
		Type:        FileTypeFile,
		FilePath:    filePath,
		IsPermanent: isPermanent,
		Owner:       owner,
		Size:        written,
		StorageTime: time.Now(),
		FileHash:    fileHash,
	}
	// 添加到 DataSpecIndex
	key := getKey(fileName, tag)
	dsl.DataSpecIndex[key] = d

	return &d, nil
}

func (dsl *DataSpecList) SaveFolder(dest, fileName, tag, owner string, isPermanent bool) (*DataSpec, error) {
	dsl.mutex.Lock()
	defer dsl.mutex.Unlock()
	// 生成文件路径
	filePath := generateFilePath(dest, fileName, tag)
	// 检查文件是否存在
	if fileExists(filePath) {
		return nil, fmt.Errorf("folder already exists: %s_%s", fileName, tag)
	}
	// 创建文件夹对象
	d := DataSpec{
		FileName:    fileName,
		Tag:         tag,
		Type:        FileTypeFolder,
		FilePath:    filePath,
		IsPermanent: isPermanent,
		Owner:       owner,
		Size:        0,
		StorageTime: time.Now(),
		FileHash:    "",
	}
	// 添加到 DataSpecIndex
	key := getKey(fileName, tag)
	dsl.DataSpecIndex[key] = d
	return &d, nil
}

// LoadFile 从指定目录加载文件，并更新 DataSpecIndex
func (dsl *DataSpecList) LoadFile(fileName, tag string) (file *os.File, err error) {
	dsl.mutex.Lock()
	defer dsl.mutex.Unlock()

	key := getKey(fileName, tag)
	dataSpec, exists := dsl.DataSpecIndex[key]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	// 加载文件
	filePath := dataSpec.FilePath
	file, err = os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// GetDataSpec 根据文件路径获取 DataSpec
func (dsl *DataSpecList) GetDataSpec(fileName, tag string) (*DataSpec, error) {
	dsl.mutex.Lock()
	defer dsl.mutex.Unlock()

	// 查询文件是否在 DataSpecIndex 中
	key := getKey(fileName, tag)
	dataSpec, exists := dsl.DataSpecIndex[key]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	return &dataSpec, nil
}

// GetList 列出所有 DataSpec
func (dsl *DataSpecList) GetList() []DataSpec {
	dsl.mutex.Lock()
	defer dsl.mutex.Unlock()
	var list []DataSpec
	for _, dataSpec := range dsl.DataSpecIndex {
		list = append(list, dataSpec)
	}
	return list
}

// DeleteFile 删除文件并更新 DataSpecIndex
func (dsl *DataSpecList) DeleteFile(fileName, tag string) error {
	dsl.mutex.Lock()
	defer dsl.mutex.Unlock()

	// 查询文件是否在 DataSpecIndex 中
	key := getKey(fileName, tag)
	d, exists := dsl.DataSpecIndex[key]
	if !exists {
		return fmt.Errorf("file not found: %s", key)
	}
	// 删除文件
	filePath := d.FilePath
	if err := os.Remove(filePath); err != nil {
		return err
	}
	delete(dsl.DataSpecIndex, key)
	return nil
}

// UpdateFile 更新文件并更新 DataSpecIndex
func (dsl *DataSpecList) UpdateFile(dest, fileName, tag, owner string, isPermanent bool, data io.Reader) (*DataSpec, error) {
	dsl.mutex.Lock()
	defer dsl.mutex.Unlock()
	// 查询文件是否在 DataSpecIndex 中
	key := getKey(fileName, tag)
	_, exists := dsl.DataSpecIndex[key]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	// 更新文件
	// 删除旧文件
	err := dsl.DeleteFile(fileName, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old file: %v", err)
	}
	return dsl.SaveFile(dest, fileName, tag, owner, isPermanent, data)
}
