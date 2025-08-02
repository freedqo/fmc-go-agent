package router

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller"
	"github.com/gin-gonic/gin"
)

func BindRoute(g *gin.Engine, c controller.If) {
	// 注入 用户中心Tokenid的校验
	//g.Use(c.MidInvalidTokenIdByUserCenter())

	// openai
	openai := g.Group("/openai")
	openaiv1 := openai.Group("/v1")
	openaiv1.GET("models", c.OpenAi().V1().Models)

	// openai/chat
	chat := openaiv1.Group("/chat")
	chat.POST("completions", c.OpenAi().V1().Chat().Completions)

	// session
	session := g.Group("/session")
	session.POST("creatSession", c.Session().CreatSession)
	session.POST("userSessionList", c.Session().UserSessionList)
	session.POST("sessionChatLogList", c.Session().SessionChatLogList)
	session.DELETE("deleteSessions", c.Session().DeleteSessions)
	session.DELETE("deleteChatLogs", c.Session().DeleteChatLogs)
	session.POST("querySessionChatLogsByUser", c.Session().QuerySessionChatLogsByUser)

	// knowdb
	knowdb := g.Group("/knowdb")
	knowdb.GET("/files", c.KnowDb().GetFileList)
	knowdb.DELETE("/files", c.KnowDb().DeleteFiles)
	knowdb.GET("/files/download", c.KnowDb().DownloadFile)
	knowdb.POST("/files", c.KnowDb().UploadFiles)

	prompt := g.Group("/prompt")
	prompt.GET("/getPromptTemplate", c.Prompt().GetPromptTemplate)
	prompt.POST("/creat", c.Prompt().Creat)
	prompt.DELETE("/delete", c.Prompt().Delete)
	prompt.PUT("/update", c.Prompt().Update)
	prompt.POST("/query", c.Prompt().Query)

	//mcp
	mcp := g.Group("/mcp")
	mcp.Any("", c.MCP().MCP)
}
