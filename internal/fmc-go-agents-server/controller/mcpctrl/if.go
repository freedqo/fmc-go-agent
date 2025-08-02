package mcpctrl

import "github.com/gin-gonic/gin"

type If interface {
	MCP(ctx *gin.Context)
}
