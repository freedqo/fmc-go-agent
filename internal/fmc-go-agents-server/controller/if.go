package controller

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller/knowdbctrl"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller/mcpctrl"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller/openaictrl"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller/promptctrl"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller/sessionctrl"
	"github.com/gin-gonic/gin"
)

type If interface {
	OpenAi() openaictrl.If
	Session() sessionctrl.If
	KnowDb() knowdbctrl.If
	Prompt() promptctrl.If
	MCP() mcpctrl.If
	MidInvalidTokenIdByUserCenter() gin.HandlerFunc
}
