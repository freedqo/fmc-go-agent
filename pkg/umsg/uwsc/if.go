package uwsc

import (
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
)

type If interface {
	umsg.MessageAgentIf
	GetConnectState() umsg.ClientConnectState
	// SubscribeTimerEvent 注册定时事件
	SubscribeTimerEvent(fun *func())
	// UnSubscribeTimerEvent 注销定时事件
	UnSubscribeTimerEvent(fun *func())
}
