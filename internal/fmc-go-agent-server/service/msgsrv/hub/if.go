package hub

import (
	"github.com/freedqo/fmc-go-agent/pkg/commif"
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
)

type If interface {
	commif.MonitorIf
	// Publish 推送消息(向外)
	Publish(msg *umsg.UMsg)
	// SubscribeRecEvent 注册接收消息的代理（向内）
	SubscribeRecEvent(fun *umsg.RecEventFunc)
	// UnSubscribeRecEvent 注销接收消息的代理
	UnSubscribeRecEvent(fun *umsg.RecEventFunc)
	// GetMsgHubIf 获取消息中心总线接口
	GetMsgHubIf(name string) umsg.MessageHubIf
}
