package controller

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller/knowdbctrl"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller/mcpctrl"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller/openaictrl"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller/promptctrl"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller/sessionctrl"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service"
	"github.com/freedqo/fmc-go-agents/pkg/webapp"
	"github.com/gin-gonic/gin"
	"net/http"
)

func New(service service.If) If {
	return &Controller{
		openai:  openaictrl.New(service),
		session: sessionctrl.New(service),
		knowdb:  knowdbctrl.New(service.KnowDb()),
		prompt:  promptctrl.New(service.Prompt()),
		mcp:     mcpctrl.New(service.MCP()),
		service: service,
	}
}

type Controller struct {
	service service.If
	openai  openaictrl.If
	session sessionctrl.If
	knowdb  knowdbctrl.If
	prompt  promptctrl.If
	mcp     mcpctrl.If
}

func (ctrl *Controller) MCP() mcpctrl.If {
	return ctrl.mcp
}

func (ctrl *Controller) Prompt() promptctrl.If {
	return ctrl.prompt
}

func (ctrl *Controller) KnowDb() knowdbctrl.If {
	return ctrl.knowdb
}

func (ctrl *Controller) MidInvalidTokenIdByUserCenter() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenId := c.GetHeader("Tokenid")
		if tokenId == "" {
			res := &webapp.Response{
				Code:    http.StatusUnauthorized,
				Message: "Tokenid is empty",
				Data:    nil,
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, res)
			return
		}
		err := ctrl.service.MidInvalidTokenIdByUserCenter(c.Request.Context(), tokenId)
		if err != nil {
			res := &webapp.Response{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
				Data:    nil,
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, res)
			return
		}
		c.Next()
	}
}

func (ctrl *Controller) Session() sessionctrl.If {
	return ctrl.session
}

func (ctrl *Controller) OpenAi() openaictrl.If {
	return ctrl.openai
}

var _ If = &Controller{}
