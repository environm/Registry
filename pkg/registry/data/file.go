package data

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type FileType string

const (
	FileTypeFile   FileType = "file"
	FileTypeFolder FileType = "folder"
)

type File struct {
	FileName    string    `json:"fileName"`
	Tag         string    `json:"tag"`
	Type        FileType  `json:"type"`
	FilePath    string    `json:"filePath"`
	Size        int64     `json:"size"`
	UploadTime  time.Time `json:"uploadTime"`
	IsPermanent bool      `json:"isPermanent"` // 新增变量，用于标识文件是否需要永久存储
}

func NewFile(fileName, tag string, fileType FileType, isPermanent bool) *File {
	return &File{
		FileName:    fileName,
		Tag:         tag,
		Type:        fileType,
		FilePath:    "",
		Size:        0,
		UploadTime:  time.Now(),
		IsPermanent: isPermanent,
	}
}

func generateFilePath(dest, fileName, tag string) string {
	return filepath.Join(dest, fmt.Sprintf("%s_%s", fileName, tag))
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func (f *File) SaveFile(dest string, data io.Reader) (err error) {
	// 确保目录存在
	err = os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return err
	}
	// 构造目标文件路径
	filePath := generateFilePath(dest, f.FileName, f.Tag)
	// 创造目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()
	written, err := io.Copy(dst, data)
	if err != nil {
		return err
	}

	// 更新 File 的字段
	f.FilePath = filePath
	f.Size = written
	f.UploadTime = time.Now()

	return nil
}

func (f *File) LoadFile(dest string) (file *os.File, err error) {
	filePath := generateFilePath(dest, f.FileName, f.Tag)
	// 检查文件是否存在
	if !fileExists(filePath) {
		return nil, fmt.Errorf("file not found: %s_%s", f.FileName, f.Tag)
	}
	file, err = os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (f *File) DeleteFile(dest string) error {
	filePath := generateFilePath(dest, f.FileName, f.Tag)
	if !fileExists(filePath) {
		return fmt.Errorf("file not found: %s_%s", f.FileName, f.Tag)
	}
	err := os.Remove(filePath)
	return err
}

func (f *File) UpdateFile(dest string, data io.Reader) error {
	filePath := generateFilePath(dest, f.FileName, f.Tag)
	// 检查文件是否存在
	if !fileExists(filePath) {
		return fmt.Errorf("file not found: %s_%s", f.FileName, f.Tag)
	}
	// 删除旧文件
	err := f.DeleteFile(dest)
	if err != nil {
		return fmt.Errorf("failed to delete old file: %v", err)
	}

	// 保存新文件
	return f.SaveFile(dest, data)
}

func (f *File) IsExists(dest string) bool {
	filePath := generateFilePath(dest, f.FileName, f.Tag)
	return fileExists(filePath)
}

func (f *File) SetPermanent(isPermanent bool) {
	f.IsPermanent = isPermanent
}

// GetIsPermanent 根据文件名和标签获取文件的 IsPermanent 值
func (f *File) GetIsPermanent() bool {
	return f.IsPermanent
}
