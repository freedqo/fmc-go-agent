package ext

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal/ext/usercenter"
)

type If interface {
	UserCenter() (usercenter.If, error)
}
