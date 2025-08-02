package hub

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/msgsrv/hub/mqtt"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/store/log"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	"sync"
)

// New 创建消息集线器
// 入参：cfg *config.DisplayServerAppCfgFile 配置文件
// 返回： If 消息集线器接口
func New(cfg *config.Config) If {
	hub := Hub{
		bus: make(map[string]fmsg.MessageHubIf, 0),
		//wss: make(map[string]fmsg.MessageHubIf, 0),
		mqt: make(map[string]fmsg.MessageHubIf, 0),
	}
	//// 注入wss消息实例
	//if cfg.Msg.MainWss != nil && cfg.Msg.MainWss.Enable {
	//	wss3 := mainsrvws.New(cfg.Msg.MainWss)
	//	hub.wss[wss3.GetName()] = wss3
	//}

	// 注入mqtt消息实例
	if cfg.Msg.Mqtt != nil && cfg.Msg.Mqtt.Enable {
		mqt1 := mqtt.New(cfg.Msg.Mqtt)
		hub.mqt[mqt1.GetName()] = mqt1
	}
	//// 注入bus消息实例
	//for k, v := range hub.wss {
	//	_, ok := hub.bus[k]
	//	if !ok {
	//		hub.bus[k] = v
	//	} else {
	//		panic("重复注入消息集线器实例:" + k)
	//	}
	//}
	for k, v := range hub.mqt {
		_, ok := hub.bus[k]
		if !ok {
			hub.bus[k] = v
		} else {
			panic("重复注入消息集线器实例:" + k)
		}
	}
	return &hub
}

type Hub struct {
	ctx    context.Context
	cancel context.CancelFunc
	bus    map[string]fmsg.MessageHubIf
	//wss    map[string]fmsg.MessageHubIf
	mqt map[string]fmsg.MessageHubIf
}

// GetMsgHubIf 获取wss消息集线器
// 入参：name string 消息集线器名称
// 返回：fmsg.MessageHubIf 消息集线器接口
func (h *Hub) GetMsgHubIf(name string) fmsg.MessageHubIf {
	v, ok := h.bus[name]
	if ok {
		return v
	}
	return nil
}

// SubscribeRecEvent 订阅消息接收事件
// 入参：fun *fmsg.RecEventFunc 消息接收事件函数
// 返回：无
func (h *Hub) SubscribeRecEvent(fun *fmsg.RecEventFunc) {
	for _, v := range h.bus {
		v.SubscribeRecEvent(fun)
	}
}

// UnSubscribeRecEvent 取消订阅消息接收事件
// 入参：fun *fmsg.RecEventFunc 消息接收事件函数
// 返回：无
func (h *Hub) UnSubscribeRecEvent(fun *fmsg.RecEventFunc) {
	for _, v := range h.bus {
		v.UnSubscribeRecEvent(fun)
	}
}

// Start 启动消息集线器
// 入参：ctx context.Context 上下文
// 返回：done <-chan struct{} 退出信号
// 返回：err error 错误信息
func (h *Hub) Start(ctx context.Context) (done <-chan struct{}, err error) {
	d := make(chan struct{})
	lctx, cancel := context.WithCancel(ctx)
	h.ctx = ctx
	h.cancel = cancel
	var once sync.Once
	for _, v := range h.bus {
		d1, err1 := v.Start(lctx)
		if err1 != nil {
			cancel()
			return nil, err1
		}
		go func(c <-chan struct{}) {
			defer func() {
				once.Do(func() {
					close(d)
				})
			}()
			select {
			case <-c:
				{
					cancel()
					return
				}
			case <-lctx.Done():
				{
					return
				}
			}
		}(d1)
	}
	return d, nil
}

// Stop 停止消息集线器
// 入参：无
// 返回：err error 错误信息
func (h *Hub) Stop() error {
	if h.cancel == nil {
		return nil
	}
	h.cancel()
	h.cancel = nil
	return nil
}

// RestStart 重启消息集线器
// 入参：无
// 返回：done <-chan struct{} 退出信号
// 返回：err error 错误信息
func (h *Hub) RestStart() (done <-chan struct{}, err error) {
	err = h.Stop()
	if err != nil {
		return nil, err
	}
	return h.Start(h.ctx)
}

// Publish 推送消息
// 入参：msg *fmsg.UMsg 消息
// 返回：无
func (h *Hub) Publish(msg *fmsg.UMsg) {
	if msg.OutType == nil || len(msg.OutType) <= 0 {
		log.SysLog().Warnf("推送异常:%v", msg.String("未知输出类型消息"))
		return
	}
	for _, to := range msg.OutType {
		switch to {
		//case fmsg.ToWsServer:
		//	{
		//		for _, v := range h.wss {
		//			v.Publish(msg)
		//			log.SysLog().Infof("mcp publish msg: %v", msg.String("推送到wss的客户端"))
		//		}
		//		break
		//	}
		case fmsg.ToMqtt:
			{
				for _, v := range h.mqt {
					v.Publish(msg)
					log.SysLog().Infof("mcp publish msg: %v", msg.String("推送到mqtt的服务器"))
				}
				break
			}
		default:
			{
				log.SysLog().Warnf("推送异常:%v", msg.String("未知输出类型消息"))
				break
			}
		}
	}
}

var _ If = &Hub{}
