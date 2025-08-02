package uwsc

import (
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
)

type If interface {
	fmsg.MessageAgentIf
	GetConnectState() fmsg.ClientConnectState
	// SubscribeTimerEvent 注册定时事件
	SubscribeTimerEvent(fun *func())
	// UnSubscribeTimerEvent 注销定时事件
	UnSubscribeTimerEvent(fun *func())
}
