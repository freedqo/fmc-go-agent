package knowdbsrv

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/knowdbm"
	"mime/multipart"
)

type If interface {
	GetFileList(req knowdbm.GetFileListReq) (*knowdbm.GetFileListResp, error)
	DeleteFiles(req knowdbm.DeleteFilesReq) (res *knowdbm.DeleteFilesResp, err error)
	Download(file string) (string, error)
	UploadFiles(files []*multipart.FileHeader) error
}
