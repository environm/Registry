package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
)

// TODO:

type UploadHandler struct {
	//
	DataPath string
	//FileMapping *data.FileMapping
	DataSpecList *data.DataSpecList

	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *UploadHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewUploadHandler(dataPath string, dataSpecList *data.DataSpecList) *UploadHandler {
	dh := &UploadHandler{
		DataPath:     dataPath,
		DataSpecList: dataSpecList,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &UploadHandler{}

func (d *UploadHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
			return
		}

		fileName, tag, owner, fileType, err := utils.GetFileParams(r, "v1.0.0")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting file parameters: %v", err), http.StatusBadRequest)
			return
		}

		// 获取是否需要永久存储的参数
		//isPermanent := utils.GetQueryParamCaseInsensitive(param, "isPermanent") == "true"

		// 通过 upload 上传的文件默认为持久化存储， receive 收到的文件默认为临时存储(区别在于 下载之后是否删除)

		switch fileType {
		case "folder", "completion":
			utils.ReceiveDir(w, r, d.DataPath)
			_, err := d.DataSpecList.SaveFolder(d.DataPath, fileName, tag, owner, true)
			//err := d.FileMapping.SaveFolder(d.DataPath, fileName, tag, true)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to save folder: %v", err), http.StatusInternalServerError)
				return
			}
		case "file":
			//err := d.FileMapping.SaveFile(d.DataPath, fileName, tag, true, r.Body)
			_, err := d.DataSpecList.SaveFile(d.DataPath, fileName, tag, owner, true, r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		//w.WriteHeader(http.StatusCreated)
		w.Write([]byte("File uploaded successfully"))
	}
}
