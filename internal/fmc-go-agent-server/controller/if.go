package controller

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/controller/knowdbctrl"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/controller/openaictrl"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/controller/promptctrl"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/controller/sessionctrl"
	"github.com/gin-gonic/gin"
)

type If interface {
	OpenAi() openaictrl.If
	Session() sessionctrl.If
	KnowDb() knowdbctrl.If
	Prompt() promptctrl.If
	MidInvalidTokenIdByUserCenter() gin.HandlerFunc
}
