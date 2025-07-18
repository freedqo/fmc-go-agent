package handler

import (
	"context"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/urecover"
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
)

func New() If {
	ctx := context.Background()
	return &Handler{
		ctx: ctx,
	}
}

type Handler struct {
	ctx context.Context
}

func (h *Handler) HandleInMsgFunc(push umsg.PublishFunc, msg *umsg.UMsg) {
	defer urecover.HandlerRecover("总线消息处理器，处理待推送消息", nil)

	if msg == nil || msg.Msg == nil {
		return
	}

	// 处理其他具体的业务消息
	switch msg.Flag {
	case "McpServer":
		{
			push(msg)
			break
		}
	default: //无消息来源标志,不推送并打印日志
		{
			//log.SysLog().Errorf("消息:%s,无法推送,消息来源标志为空", msg.String("消息来源标志为空"))
			break
		}
	}

	return
}

var _ If = &Handler{}
