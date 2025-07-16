package mqtt

import (
	"context"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/iconsts"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/log"
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
	"github.com/freedqo/fmc-go-agent/pkg/umsg/umqt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sync"
)

func New(cfg *umqt.Option) If {
	cfg.SubTopic = make([]string, 0)
	// 确保程序内，固化的主题都订阅了
	if cfg.SubTopic != nil {
		for _, v := range iconsts.SubTopicList {
			isExit := false
			for _, topic := range cfg.SubTopic {
				if topic == v {
					isExit = true
					break
				}
			}
			if !isExit {
				cfg.SubTopic = append(cfg.SubTopic, v)
			}
		}
	} else {
		cfg.SubTopic = iconsts.SubTopicList
	}
	name := "utaimqtt-" + uuid.New().String()
	c := Client{
		mu:   sync.Mutex{},
		cfg:  cfg,
		Name: name,
		mqtt: umqt.New(name, cfg, log.SysLog()),
		connState: umsg.ClientConnectState{
			ClientID:    name,
			IsConnected: false,
		},
		recMsgFun: make(map[*umsg.RecEventFunc]*umsg.RecEventFunc),
		log:       log.SysLog(),
	}
	c.recMsgFunc = c.onOutMsg
	return &c
}

type Client struct {
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
	Name       string
	mqtt       umsg.MessageAgentIf
	cfg        *umqt.Option
	connState  umsg.ClientConnectState
	recMsgFun  map[*umsg.RecEventFunc]*umsg.RecEventFunc //接收消息回调
	recMsgFunc umsg.RecEventFunc
	log        *zap.SugaredLogger
}

func (c *Client) GetName() string {
	return c.Name
}

func (c *Client) onOutMsg(msg *umsg.UMsg) {
	c.mu.Lock()
	defer c.mu.Unlock()
	msg.Flag = c.Name
	msg.OutType = nil
	for _, fun := range c.recMsgFun {
		(*fun)(msg)
	}
}

func (c *Client) Start(ctx context.Context) (done <-chan struct{}, err error) {
	if !c.cfg.Enable {
		c.log.Infof("MQTT[%s]启动条件不满足，不启动MQTT消息服务", c.Name)
		return make(chan struct{}), err
	}
	c.mqtt.SubscribeRecEvent(&c.recMsgFunc)
	return c.mqtt.Start(ctx)
}

func (c *Client) Stop() error {
	if !c.cfg.Enable {
		return nil
	}
	c.mqtt.UnSubscribeRecEvent(&c.recMsgFunc)
	return c.mqtt.Stop()
}

func (c *Client) RestStart() (done <-chan struct{}, err error) {
	if !c.cfg.Enable {
		c.log.Infof("MQTT[%s]启动条件不满足，不启动MQTT消息服务", c.Name)
		return make(chan struct{}), err
	}
	return c.mqtt.RestStart()
}

func (c *Client) GetConnectState() umsg.ClientConnectState {
	if !c.cfg.Enable {
		return umsg.ClientConnectState{
			ClientID:    c.Name,
			IsConnected: false,
		}
	}
	return c.mqtt.GetConnectState()
}

func (c *Client) Publish(msg *umsg.UMsg) {
	if !c.cfg.Enable {
		return
	}
	c.mqtt.Publish(msg)
}

func (c *Client) SubscribeRecEvent(fun *umsg.RecEventFunc) {
	if !c.cfg.Enable {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recMsgFun[fun] = fun
}

func (c *Client) UnSubscribeRecEvent(fun *umsg.RecEventFunc) {
	if !c.cfg.Enable {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.recMsgFun, fun)
}

func (c *Client) FrontHandleMessage(msg *umsg.UMsg) {
	// 不是ws标准输入,暂不处理
}

var _ If = &Client{}
