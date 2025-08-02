package mqtt

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/iconsts"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/store/log"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg/umqt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sync"
)

func New(cfg *umqt.Option) If {
	if cfg.SubTopic == nil {
		cfg.SubTopic = make([]string, 0)
	}
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
		connState: fmsg.TConnStatus{
			Name:  name,
			State: false,
		},
		recMsgFun: make(map[*fmsg.RecEventFunc]*fmsg.RecEventFunc),
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
	mqtt       fmsg.MessageAgentIf
	cfg        *umqt.Option
	connState  fmsg.TConnStatus
	recMsgFun  map[*fmsg.RecEventFunc]*fmsg.RecEventFunc //接收消息回调
	recMsgFunc fmsg.RecEventFunc
	log        *zap.SugaredLogger
}

func (c *Client) GetName() string {
	return c.Name
}

func (c *Client) onOutMsg(msg *fmsg.UMsg) {
	c.mu.Lock()
	defer c.mu.Unlock()
	msg.Sour = c.Name
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

func (c *Client) GetConnectState() fmsg.TConnStatus {
	if !c.cfg.Enable {
		return fmsg.TConnStatus{
			Name:  c.Name,
			State: false,
		}
	}
	return c.mqtt.GetConnectState()
}

func (c *Client) Publish(msg *fmsg.UMsg) {
	if !c.cfg.Enable {
		return
	}
	c.mqtt.Publish(msg)
}

func (c *Client) SubscribeRecEvent(fun *fmsg.RecEventFunc) {
	if !c.cfg.Enable {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recMsgFun[fun] = fun
}

func (c *Client) UnSubscribeRecEvent(fun *fmsg.RecEventFunc) {
	if !c.cfg.Enable {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.recMsgFun, fun)
}

func (c *Client) FrontHandleMessage(msg *fmsg.UMsg) {
	// 不是ws标准输入,暂不处理
}

var _ If = &Client{}
