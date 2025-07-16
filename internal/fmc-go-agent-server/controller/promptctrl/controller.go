package promptctrl

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/promptm"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/promptsrv"
	"github.com/freedqo/fmc-go-agent/pkg/webapp"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Controller struct {
	service promptsrv.If
}

func New(s promptsrv.If) If {
	return &Controller{service: s}
}

// GetPromptTemplate 获取提示词模板
//
//	@Summary		获取系统提示词模板
//	@Description	获取系统提示词模板
//	@Tags			提示词管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid	header		string	true	"Tokenid 用户登录令牌"
//	@Success		200		{object}	webapp.Response{data=promptm.GetPromptTemplateResp}
//	@Failure		400		{object}	webapp.Response
//	@Failure		500		{object}	webapp.Response
//	@Router			/prompt/getPromptTemplate [get]
func (c *Controller) GetPromptTemplate(ctx *gin.Context) {
	resp, err := c.service.GetPromptTemplate(ctx.Request.Context(), struct{}{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, webapp.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	ctx.JSON(http.StatusOK, webapp.Response{
		Code:    http.StatusOK,
		Message: "",
		Data:    resp,
	})
}

// Creat 添加提示词
//
//	@Summary		添加提示词
//	@Description	添加提示词
//	@Tags			提示词管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid	header		string				true	"Tokenid 用户登录令牌"
//	@Param			request	body		promptm.CreatReq	true	"请求参数"
//	@Success		200		{object}	webapp.Response{data=promptm.CreatResp}
//	@Failure		400		{object}	webapp.Response
//	@Failure		500		{object}	webapp.Response
//	@Router			/prompt/creat [post]
func (c *Controller) Creat(ctx *gin.Context) {
	var req promptm.CreatReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, webapp.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	resp, err := c.service.Creat(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, webapp.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	ctx.JSON(http.StatusOK, webapp.Response{
		Code:    http.StatusOK,
		Message: "",
		Data:    resp,
	})
}

// Delete 删除提示词
//
//	@Summary		删除提示词
//	@Description	删除提示词
//	@Tags			提示词管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid				header		string				true	"Tokenid 用户登录令牌"
//	@Param			promptm.DeleteReq	body		promptm.DeleteReq	true	"请求参数"
//	@Success		200					{object}	webapp.Response{data=promptm.DeleteResp}
//	@Failure		400					{object}	webapp.Response
//	@Failure		500					{object}	webapp.Response
//	@Router			/prompt/Delete [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	req := promptm.DeleteReq{}
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, webapp.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	resp, err := c.service.Delete(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, webapp.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	ctx.JSON(http.StatusOK, webapp.Response{
		Code:    http.StatusOK,
		Message: "",
		Data:    resp,
	})
}

// Query 查询提示词
//
//	@Summary		查询提示词
//	@Description	查询提示词
//	@Tags			提示词管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid				header		string				true	"Tokenid 用户登录令牌"
//	@Param			promptm.QueryReq	body		promptm.QueryReq	false	"请求参数"
//	@Success		200					{object}	webapp.Response{data=promptm.QueryResp}
//	@Failure		400					{object}	webapp.Response
//	@Failure		500					{object}	webapp.Response
//	@Router			/prompt/query [post]
func (c *Controller) Query(ctx *gin.Context) {
	req := promptm.QueryReq{}
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, webapp.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	resp, err := c.service.Query(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, webapp.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	ctx.JSON(http.StatusOK, webapp.Response{
		Code:    http.StatusOK,
		Message: "",
		Data:    resp,
	})
}

// Update 修改提示词
//
//	@Summary		修改提示词
//	@Description	修改提示词
//	@Tags			提示词管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid				header		string				true	"Tokenid 用户登录令牌"
//	@Param			promptm.UpdateReq	body		promptm.UpdateReq	true	"请求参数"
//	@Success		200					{object}	webapp.Response{data=promptm.UpdateResp}
//	@Failure		400					{object}	webapp.Response
//	@Failure		500					{object}	webapp.Response
//	@Router			/prompt/update [put]
func (c *Controller) Update(ctx *gin.Context) {
	var req promptm.UpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, webapp.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	resp, err := c.service.Update(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, webapp.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	ctx.JSON(http.StatusOK, webapp.Response{
		Code:    http.StatusOK,
		Message: "",
		Data:    resp,
	})
}
