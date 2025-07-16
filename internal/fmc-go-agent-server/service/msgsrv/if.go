package msgsrv

import (
	"github.com/freedqo/fmc-go-agent/pkg/commif"
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
	"time"
)

type If interface {
	commif.MonitorIf
	// Publish 推送待处理的消息（写入处理器）
	Publish(msg *umsg.UMsg)

	// pushMsg 推送消息（推送出去）
	push(msg *umsg.UMsg)

	// WaitingResMsg 等待响应消息
	WaitingResMsg(operateID string, waitTime time.Duration) (*umsg.UMsg, error)

	// Subscribe 注册接收消息的代理
	Subscribe(fun *umsg.HandlerInMsgFun)
	// UnSubscribe 注销接收消息的代理
	UnSubscribe(fun *umsg.HandlerInMsgFun)

	GetMsgHubIf(name string) umsg.MessageHubIf
}
