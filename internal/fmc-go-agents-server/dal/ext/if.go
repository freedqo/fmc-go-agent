package ext

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/ext/knowledge"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/ext/usercenter"
)

type If interface {
	UserCenter() (usercenter.If, error)
	Knowledge() (knowledge.If, error)
}
