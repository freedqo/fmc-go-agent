package umsg

import (
	"github.com/freedqo/fmc-go-agent/pkg/commif"
)

// MessageAgentIf 消息服务器或者客户端的代理接口
type MessageAgentIf interface {
	commif.MonitorIf
	// Publish 推送消息(向外)
	Publish(msg *UMsg)
	// SubscribeRecEvent 注册接收消息的代理（向内）
	SubscribeRecEvent(fun *RecEventFunc)
	// UnSubscribeRecEvent 注销接收消息的代理
	UnSubscribeRecEvent(fun *RecEventFunc)
	// GetConnectState 获取连接状态
	GetConnectState() ClientConnectState
}

// MessageHubIf 消息代理接入处理器接口
type MessageHubIf interface {
	MessageAgentIf
	// FrontHandleMessage 处理前置消息处理器，归一化外部消息为内部可识别的消息(不做业务逻辑)
	FrontHandleMessage(msg *UMsg)
	GetName() string
}
