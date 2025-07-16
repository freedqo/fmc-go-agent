package sessionctrl

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/sessionm"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service"
	"github.com/freedqo/fmc-go-agent/pkg/webapp"
	"github.com/gin-gonic/gin"
	"net/http"
)

func New(service service.If) If {
	return &Controller{
		service: service,
	}
}

type Controller struct {
	service service.If
}

// CreatSession 创建会话
//
//	@Summary		创建会话
//	@Description	创建会话
//	@Tags			会话管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid						header		string											true	"Tokenid 用户登录令牌"
//	@Param			sessionm.CreatSessionReq	body		sessionm.CreatSessionReq						true	"请求参数"
//	@Success		200							{object}	webapp.Response{data=sessionm.CreatSessionResp}	"成功响应"
//	@Router			/session/creatSession [post]
func (ctrl *Controller) CreatSession(c *gin.Context) {
	req := &sessionm.CreatSessionReq{}
	err := c.ShouldBind(req)
	if err != nil {
		c.JSON(400, &webapp.Response{
			Code:    400,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	resp, err := ctrl.service.Session().CreatSession(c.Request.Context(), *req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, webapp.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	c.JSON(200, &webapp.Response{
		Code:    200,
		Message: "",
		Data:    resp,
	})
	return
}

// UserSessionList 查询用户对话列表
//
//	@Summary		查询用户对话列表
//	@Description	查询用户对话列表
//	@Tags			会话管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid						header		string												true	"Tokenid 用户登录令牌"
//	@Param			sessionm.UserSessionListReq	body		sessionm.UserSessionListReq							true	"请求参数"
//	@Success		200							{object}	webapp.Response{data=sessionm.UserSessionListResp}	"成功响应"
//	@Router			/session/userSessionList [post]
func (ctrl *Controller) UserSessionList(c *gin.Context) {
	req := &sessionm.UserSessionListReq{}
	err := c.ShouldBind(req)
	if err != nil {
		c.JSON(400, &webapp.Response{
			Code:    400,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	resp, err := ctrl.service.Session().UserSessionList(c.Request.Context(), *req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, webapp.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	c.JSON(200, &webapp.Response{
		Code:    200,
		Message: "",
		Data:    resp,
	})
	return
}

// SessionChatLogList 查询聊天内容
//
//	@Summary		查询聊天内容
//	@Description	查询聊天内容
//	@Tags			会话管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid							header		string													true	"Tokenid 用户登录令牌"
//	@Param			sessionm.SessionChatLogListReq	body		sessionm.SessionChatLogListReq							true	"请求参数"
//	@Success		200								{object}	webapp.Response{data=sessionm.SessionChatLogListResp}	"成功响应"
//	@Router			/session/sessionChatLogList [post]
func (ctrl *Controller) SessionChatLogList(c *gin.Context) {
	req := &sessionm.SessionChatLogListReq{}
	err := c.ShouldBind(req)
	if err != nil {
		c.JSON(400, &webapp.Response{
			Code:    400,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	resp, err := ctrl.service.Session().SessionChatLogList(c.Request.Context(), *req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, webapp.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	c.JSON(200, &webapp.Response{
		Code:    200,
		Message: "",
		Data:    resp,
	})
	return
}

// DeleteSessions 删除多个用户对话
//
//	@Summary		删除多个用户对话
//	@Description	删除多个用户对话
//	@Tags			会话管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid						header		string												true	"Tokenid 用户登录令牌"
//	@Param			sessionm.DeleteSessionsReq	body		sessionm.DeleteSessionsReq							true	"请求参数"
//	@Success		200							{object}	webapp.Response{data=sessionm.DeleteSessionsResp}	"成功响应"
//	@Router			/session/deleteSessions [DELETE]
func (ctrl *Controller) DeleteSessions(c *gin.Context) {
	req := &sessionm.DeleteSessionsReq{}
	err := c.ShouldBind(req)
	if err != nil {
		c.JSON(400, &webapp.Response{
			Code:    400,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	resp, err := ctrl.service.Session().DeleteSessions(c.Request.Context(), *req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, webapp.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	c.JSON(200, &webapp.Response{
		Code:    200,
		Message: "",
		Data:    resp,
	})
	return
}

// DeleteChatLogs 删除多条对话记录
//
//	@Summary		删除多条对话记录
//	@Description	删除多条对话记录
//	@Tags			会话管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid						header		string												true	"Tokenid 用户登录令牌"
//	@Param			sessionm.DeleteChatLogsReq	body		sessionm.DeleteChatLogsReq							true	"请求参数"
//	@Success		200							{object}	webapp.Response{data=sessionm.DeleteChatLogsResp}	"成功响应"
//	@Router			/session/deleteChatLogs [DELETE]
func (ctrl *Controller) DeleteChatLogs(c *gin.Context) {
	req := &sessionm.DeleteChatLogsReq{}
	err := c.ShouldBind(req)
	if err != nil {
		c.JSON(400, &webapp.Response{
			Code:    400,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	resp, err := ctrl.service.Session().DeleteChatLogs(c.Request.Context(), *req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, webapp.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	c.JSON(200, &webapp.Response{
		Code:    200,
		Message: "",
		Data:    resp,
	})
	return
}

// QuerySessionChatLogsByUser 根据用户信息和提示词类型获取唯一会话记录
//
//	@Summary		根据用户信息和提示词类型获取唯一会话记录
//	@Description	根据用户信息和提示词类型获取唯一会话记录
//	@Tags			会话管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid									header		string															true	"Tokenid 用户登录令牌"
//	@Param			sessionm.QuerySessionChatLogsByUserReq	body		sessionm.QuerySessionChatLogsByUserReq							true	"请求参数"
//	@Success		200										{object}	webapp.Response{data=sessionm.QuerySessionChatLogsByUserResp}	"成功响应"
//	@Router			/session/querySessionChatLogsByUser [post]
func (ctrl *Controller) QuerySessionChatLogsByUser(c *gin.Context) {
	req := &sessionm.QuerySessionChatLogsByUserReq{}
	err := c.ShouldBind(req)
	if err != nil {
		c.JSON(400, &webapp.Response{
			Code:    400,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	resp, err := ctrl.service.Session().QuerySessionChatLogsByUser(c.Request.Context(), *req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, webapp.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	c.JSON(200, &webapp.Response{
		Code:    200,
		Message: "",
		Data:    resp,
	})
	return
}

var _ If = &Controller{}
