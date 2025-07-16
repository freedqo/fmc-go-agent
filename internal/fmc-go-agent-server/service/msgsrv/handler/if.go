package handler

import (
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
)

type If interface {
	HandleInMsgFunc(push umsg.PublishFunc, msg *umsg.UMsg)
}
