package v1ctrl

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/controller/openaictrl/v1ctrl/chatctrl"
	"github.com/gin-gonic/gin"
)

type If interface {
	Chat() chatctrl.If
	Models(c *gin.Context)
}
