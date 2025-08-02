package knowdbctrl

import "github.com/gin-gonic/gin"

type If interface {
	GetFileList(ctx *gin.Context)
	DeleteFiles(ctx *gin.Context)
	DownloadFile(ctx *gin.Context)
	UploadFiles(ctx *gin.Context)
}
