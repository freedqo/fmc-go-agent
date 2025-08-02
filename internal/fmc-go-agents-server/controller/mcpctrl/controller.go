package mcpctrl

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/mcpsrv"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service mcpsrv.If
}

func New(service mcpsrv.If) If {
	return &Controller{
		service: service,
	}
}

func (c *Controller) MCP(ctx *gin.Context) {
	c.service.ServeHTTP(ctx.Writer, ctx.Request)
}
