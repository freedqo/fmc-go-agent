package knowdbctrl

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/knowdbsrv"
	"github.com/freedqo/fmc-go-agents/pkg/webapp"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/knowdbm"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service knowdbsrv.If
}

func New(service knowdbsrv.If) If {
	return &Controller{
		service: service,
	}
}

// GetFileList godoc
//
//	@Summary		获取文件列表
//	@Description	根据条件获取文件列表
//	@Tags			知识库管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid		header		string	true	"Tokenid 用户登录令牌"
//	@Param			type		query		string	false	"文件类型"
//	@Param			page		query		int		false	"页码"	default(1)
//	@Param			pageSize	query		int		false	"每页数量"	default(20)
//	@Success		200			{object}	webapp.Response{data=knowdbm.GetFileListResp}
//	@Failure		400			{object}	webapp.Response
//	@Failure		500			{object}	webapp.Response
//	@Router			/knowdb/files [get]
func (c *Controller) GetFileList(ctx *gin.Context) {
	req := knowdbm.GetFileListReq{}
	res := webapp.Response{
		Code:    200,
		Message: "",
		Data:    nil,
	}
	// 可选查询参数
	if fileType := ctx.Query("type"); fileType != "" {
		req.Type = fileType
	}

	if pageStr := ctx.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = int32(page)
		}
	}

	if pageSizeStr := ctx.Query("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = int32(pageSize)
		}
	}
	resp, err := c.service.GetFileList(req)
	if err != nil {
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}
	res.Data = resp
	ctx.JSON(http.StatusOK, res)
}

// DeleteFiles godoc
//
//	@Summary		删除文件
//	@Description	根据文件ID删除文件
//	@Tags			知识库管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid	header		string					true	"Tokenid 用户登录令牌"
//	@Param			ids		body		knowdbm.DeleteFilesReq	true	"文件ID列表"
//	@Success		200		{object}	webapp.Response{data=knowdbm.DeleteFilesResp}
//	@Failure		400		{object}	webapp.Response
//	@Failure		500		{object}	webapp.Response
//	@Router			/knowdb/files [delete]
func (c *Controller) DeleteFiles(ctx *gin.Context) {
	var req knowdbm.DeleteFilesReq
	res := webapp.Response{
		Code:    200,
		Message: "",
		Data:    nil,
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	resp, err := c.service.DeleteFiles(req)
	if err != nil {
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}
	res.Data = resp
	ctx.JSON(http.StatusOK, res)
}

// DownloadFile godoc
//
//	@Summary		下载文件
//	@Description	根据文件路径下载文件
//	@Tags			知识库管理
//	@Accept			json
//	@Produce		octet-stream
//	@Param			Tokenid	header		string	true	"Tokenid 用户登录令牌"
//	@Param			id		query		string	true	"文件Id"
//	@Success		200		{file}		file
//	@Failure		400		{object}	webapp.Response
//	@Failure		404		{object}	webapp.Response
//	@Failure		500		{object}	webapp.Response
//	@Router			/knowdb/files/download [get]
func (c *Controller) DownloadFile(ctx *gin.Context) {
	res := webapp.Response{
		Code:    200,
		Message: "",
		Data:    nil,
	}
	id := ctx.Query("id")
	if id == "" {
		res.Code = http.StatusBadRequest
		res.Message = "文件名称不能为空"
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	realPath, err := c.service.Download(id)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			res.Code = http.StatusNotFound
			res.Message = err.Error()
			ctx.JSON(http.StatusNotFound, res)
		} else {
			res.Code = http.StatusInternalServerError
			res.Message = err.Error()
			ctx.JSON(http.StatusInternalServerError, res)
		}
		return
	}
	// 获取文件名
	filename := ctx.Query("filename")
	if filename == "" {
		filename = filepath.Base(realPath)
	}

	ctx.FileAttachment(realPath, filename)
}

// UploadFiles godoc
//
//	@Summary		上传文件
//	@Description	上传单个或多个文件
//	@Tags			知识库管理
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			Tokenid	header		string	true	"Tokenid 用户登录令牌"
//	@Param			files	formData	array	true	"文件列表"
//	@Success		200		{object}	webapp.Response{data=map[string][]string}
//	@Failure		400		{object}	webapp.Response
//	@Failure		500		{object}	webapp.Response
//	@Router			/knowdb/files [post]
func (c *Controller) UploadFiles(ctx *gin.Context) {
	res := webapp.Response{
		Code:    200,
		Message: "",
		Data:    nil,
	}
	// 单文件上传
	file, err := ctx.FormFile("file")
	if err == nil {
		err := c.service.UploadFiles([]*multipart.FileHeader{file})
		if err != nil {
			res.Code = http.StatusInternalServerError
			res.Message = err.Error()
			ctx.JSON(http.StatusInternalServerError, res)
			return
		}
		res.Code = http.StatusOK
		res.Message = "文件上传成功"
		res.Data = map[string][]string{"filename": {file.Filename}}
		ctx.JSON(http.StatusOK, res)
		return
	}

	// 多文件上传
	form, err := ctx.MultipartForm()
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		res.Code = http.StatusBadRequest
		res.Message = "未上传任何文件"
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	err = c.service.UploadFiles(files)
	if err != nil {
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	filenames := make([]string, len(files))
	for i, file := range files {
		filenames[i] = file.Filename
	}
	res.Code = http.StatusOK
	res.Message = "文件上传成功"
	res.Data = map[string][]string{"filename": filenames}
	ctx.JSON(http.StatusOK, res)
}
