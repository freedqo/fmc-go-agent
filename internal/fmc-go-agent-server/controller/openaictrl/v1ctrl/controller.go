package v1ctrl

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/controller/openaictrl/v1ctrl/chatctrl"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service"
	"github.com/gin-gonic/gin"
)

func New(service service.If) If {
	return &Controller{
		service: service,
		chat:    chatctrl.New(service),
	}
}

type Controller struct {
	service service.If
	chat    chatctrl.If
}

func (ctrl *Controller) Chat() chatctrl.If {
	return ctrl.chat
}

// Models 获取可用模型列表
//
//	@Summary		获取可用模型列表
//	@Description	获取可用模型列表
//	@Tags			Openai Api 接口管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid	header		string				true	"Tokenid 用户登录令牌"
//	@Success		200		{object}	webapp.Response{}	"成功响应"
//	@Router			/openai/v1/models [get]
func (ctrl *Controller) Models(c *gin.Context) {
	models, err := ctrl.service.OpenAi().V1().Models(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, models)
}

var _ If = &Controller{}
