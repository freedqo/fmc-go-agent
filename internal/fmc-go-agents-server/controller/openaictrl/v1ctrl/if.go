package v1ctrl

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller/openaictrl/v1ctrl/chatctrl"
	"github.com/gin-gonic/gin"
)

type If interface {
	Chat() chatctrl.If
	Models(c *gin.Context)
}
