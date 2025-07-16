package sessionctrl

import (
	"github.com/gin-gonic/gin"
)

type If interface {
	CreatSession(c *gin.Context)
	UserSessionList(c *gin.Context)
	SessionChatLogList(c *gin.Context)
	DeleteSessions(c *gin.Context)
	DeleteChatLogs(c *gin.Context)
	QuerySessionChatLogsByUser(c *gin.Context)
}
