package uwss

import (
	"context"
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
	utwss2 "github.com/freedqo/fmc-go-agent/pkg/umsg/uwss/utwss"
	"github.com/freedqo/fmc-go-agent/pkg/utils"
	"go.uber.org/zap"
	"sync"
)

func NewDefaultOption(port int) *Option {
	if port <= 0 {
		port = 5001
	}
	return &Option{
		utwss2.NewDefaultOption(port),
	}
}

type Option struct {
	utwss2.Option
}

func New(name string, cfg *Option, log *zap.SugaredLogger) If {
	if log == nil {
		panic("ws server log is nil")
	}
	if cfg == nil {
		panic("ws server cfg is nil")
	}
	if name == "" {
		panic("ws server name is nil")
	}
	return &UWss{
		mu:            sync.Mutex{},
		Name:          name,
		wss:           utwss2.New(log, &cfg.Option),
		cfg:           cfg,
		outUMsgChan:   make(chan *umsg.UMsg, 1024*1),
		log:           log,
		recUMsgFunMap: make(map[*umsg.RecEventFunc]*umsg.RecEventFunc, 0),
		connState: &umsg.ClientConnectState{
			ClientID:    "",
			IsConnected: false,
		},
	}
}

type UWss struct {
	Name          string                                    // 客户端名称
	ctx           context.Context                           // 上下文
	cancel        context.CancelFunc                        // 取消函数
	mu            sync.Mutex                                // 互斥锁
	log           *zap.SugaredLogger                        // 日志接口
	connState     *umsg.ClientConnectState                  //连接状态
	outUMsgChan   chan *umsg.UMsg                           //需要推送的消息队列
	recUMsgFunMap map[*umsg.RecEventFunc]*umsg.RecEventFunc //接收消息回调

	cfg *Option        //websocket配置
	wss *utwss2.Server //utwebsocket服务端
}

func (u *UWss) GetConnectState() umsg.ClientConnectState {
	return *u.connState
}

func (u *UWss) Publish(msg *umsg.UMsg) {
	u.outUMsgChan <- msg
}

func (u *UWss) SubscribeRecEvent(fun *umsg.RecEventFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.recUMsgFunMap[fun] = fun
}

func (u *UWss) UnSubscribeRecEvent(fun *umsg.RecEventFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.recUMsgFunMap, fun)
}

func (u *UWss) Start(ctx context.Context) (done <-chan struct{}, err error) {
	lCtx, cancelFunc := context.WithCancel(ctx)
	u.ctx = ctx
	u.cancel = cancelFunc
	doneC := make(chan struct{})
	d1, err := u.wss.StartService(u.ctx)
	if err != nil {
		return nil, err
	}
	u.connState.IsConnected = true
	u.log.Infof("Ws服务器[%s],服务启动,:%d", u.Name, u.cfg.ServicePort)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				if err := recover(); err != nil {
					// 记录 panic 信息
					u.log.Errorf("Ws服务器[%s],服务异常,中止服务,错误: %s", u.Name, err)
					// 打印堆栈跟踪信息（可选）
					u.log.Errorf(utils.StackSkip(1, -1))
				}
			}
			close(doneC)
			u.connState.IsConnected = false
			u.log.Infof("Ws服务器[%s],服务停止,:%d", u.Name, u.cfg.ServicePort)
		}()
		for {
			select {
			case <-lCtx.Done():
				{
					u.log.Infof("Ws服务器[%s],请求服务停止", u.Name)
					return
				}
			case <-d1:
				{
					u.log.Warnf("Ws服务器[%s]->wss,服务中止", u.Name)
					return
				}
			case msg, ok := <-u.outUMsgChan:
				{
					if ok {
						u.log.Debugf("Ws服务器[%s],推送消息: %s", u.Name, msg.String("推送"))
						err1 := u.wss.SendMessage(*msg.Msg)
						if err1 != nil {
							u.log.Errorf("Ws服务器[%s],推送消息:%s,失败: %s", u.Name, msg.String("推送失败"), err1.Error())
						}
					}
				}
			case msg, ok := <-u.wss.ReceiveMsg:
				{
					if ok {
						if u.recUMsgFunMap != nil && len(u.recUMsgFunMap) > 0 {
							utmsg := umsg.NewUMsg(&umsg.Message{
								MessageBase: msg.BaseInfo,
								OperateData: msg.MessageData,
							}, u.Name, nil)
							u.log.Debugf("Ws服务器[%s],收到消息: %s", u.Name, utmsg.String("接收"))
							u.mu.Lock()
							for _, fun := range u.recUMsgFunMap {
								if fun != nil {
									(*fun)(utmsg)
								}
							}
							u.mu.Unlock()
						}
					}
				}
			case msg, ok := <-u.wss.ClientConnState:
				{
					if ok {
						if msg.IsConnected {
							u.log.Infof("Ws服务器[%s],客户端: %s,远方地址:%s,连接状态: %v,连接成功", u.Name, msg.ClientID, msg.RemoteAddr, msg.IsConnected)
						} else {
							u.log.Warnf("Ws服务器[%s],客户端: %s,远方地址:%s,连接状态: %v，连接断开", u.Name, msg.ClientID, msg.RemoteAddr, msg.IsConnected)
						}
					}
				}
			}
		}
	}()
	return doneC, nil
}

func (u *UWss) Stop() error {
	if u.cancel == nil {
		return nil
	}
	u.cancel()
	err := u.wss.StopService()
	if err != nil {
		return err
	}
	u.cancel = nil
	u.connState.IsConnected = false
	u.log.Infof("Ws服务器[%s],服务停止,:%d", u.Name, u.cfg.ServicePort)
	return nil
}

func (u *UWss) RestStart() (done <-chan struct{}, err error) {
	err = u.Stop()
	if err != nil {
		return nil, err
	}
	return u.Start(u.ctx)
}

var _ If = &UWss{}
