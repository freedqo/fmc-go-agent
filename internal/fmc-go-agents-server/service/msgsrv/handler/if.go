package handler

import (
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
)

type If interface {
	HandleInMsgFunc(push fmsg.PublishFunc, msg *fmsg.UMsg)
}
