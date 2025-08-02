package hub

import (
	"github.com/freedqo/fmc-go-agents/pkg/commif"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
)

type If interface {
	commif.MonitorIf
	// Publish 推送消息(向外)
	Publish(msg *fmsg.UMsg)
	// SubscribeRecEvent 注册接收消息的代理（向内）
	SubscribeRecEvent(fun *fmsg.RecEventFunc)
	// UnSubscribeRecEvent 注销接收消息的代理
	UnSubscribeRecEvent(fun *fmsg.RecEventFunc)
	// GetMsgHubIf 获取消息中心总线接口
	GetMsgHubIf(name string) fmsg.MessageHubIf
}
