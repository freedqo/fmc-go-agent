package msgsrv

import (
	"github.com/freedqo/fmc-go-agents/pkg/commif"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	"time"
)

type If interface {
	commif.MonitorIf
	// Publish 推送待处理的消息（写入处理器）
	Publish(msg *fmsg.UMsg)

	// pushMsg 推送消息（推送出去）
	push(msg *fmsg.UMsg)

	// WaitingResMsg 等待响应消息
	WaitingResMsg(operateID string, waitTime time.Duration) (*fmsg.UMsg, error)

	// Subscribe 注册接收消息的代理
	Subscribe(fun *fmsg.HandlerInMsgFun)
	// UnSubscribe 注销接收消息的代理
	UnSubscribe(fun *fmsg.HandlerInMsgFun)

	GetMsgHubIf(name string) fmsg.MessageHubIf
}
