package promptctrl

import "github.com/gin-gonic/gin"

type If interface {
	GetPromptTemplate(ctx *gin.Context)
	Creat(ctx *gin.Context)
	Delete(ctx *gin.Context)
	Query(ctx *gin.Context)
	Update(ctx *gin.Context)
}
