package chatctrl

import "github.com/gin-gonic/gin"

type If interface {
	Completions(c *gin.Context)
}
