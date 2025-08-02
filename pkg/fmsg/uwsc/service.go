package uwsc

import (
	"context"
	"encoding/json"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/socsm/socsws"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/msgm"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	utwsc2 "github.com/freedqo/fmc-go-agents/pkg/fmsg/uwsc/utwsc"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"go.uber.org/zap"
	"sync"
	"time"
)

func NewDefaultOption() *Option {
	return &Option{
		Enable:                 false,
		RemoteIP1:              "localhost",
		RemoteIP2:              "localhost",
		RemotePort:             81,
		RemotePath:             "/ws",
		WriteWaitSecond:        10,
		HandshakeTimeoutSecond: 45,
		RedialIntervalSecond:   5,
		PongWaitSecond:         60,
		IsPing:                 true,
		IsTrans:                false,
	}
}

type Option struct {
	Enable                 bool   `comment:"是否启用"`               //是否启用
	RemoteIP1              string `comment:"远方IP或主机(主)"`         //远方IP或主机(主)
	RemoteIP2              string `comment:"远方IP或主机(备)"`         //远方IP或主机(备)(如果不为空，则拨号连接时，如不拨号失败，将在RemoteIP1和RemoteIP2之间轮询拨号)
	RemotePort             int    `comment:"远方端口"`               //远方端口
	RemotePath             string `comment:"远方服务路径"`             //远方服务路径
	WriteWaitSecond        int    `comment:"写数据超时秒数"`            //写数据超时秒数
	HandshakeTimeoutSecond int    `comment:"握手超时时间秒数"`           //握手超时时间秒数
	RedialIntervalSecond   int    `comment:"重拨间隔秒数"`             //重拨间隔秒数。（即拨号失败后，间隔多长时间后再次重新拨号）
	IsPing                 bool   `comment:"是否从本客户端给服务端定时发ping"` //是否从本客户端给服务端定时发ping
	PongWaitSecond         int    `comment:"读数据超时"`              //读数据超时
	IsTrans                bool   `comment:"是否传输二进制数据"`          //是否传输二进制数据
}

// New 创建一个Ws客户端服务
// name 客户端名称
// cfg1 配置
// timeDuration 定时器时长
// log 日志
func New(name string, cfg *Option, isTrans bool, timeDuration time.Duration, log *zap.SugaredLogger) If {
	return &UWsc{
		Name: name,
		mu:   sync.Mutex{},
		cfg:  cfg,
		wsc:  utwsc2.NewClient(),
		connState: &fmsg.ClientConnectState{
			ClientID:    name,
			IsConnected: false,
		},
		outUMsgChan:   make(chan *fmsg.UMsg, 1024*1),
		log:           log,
		recUMsgFunMap: make(map[*fmsg.RecEventFunc]*fmsg.RecEventFunc, 0),
		timeDuration:  timeDuration,
		isTrans:       isTrans,
	}
}

type UWsc struct {
	Name          string                                    // 客户端名称
	ctx           context.Context                           // 上下文
	cancel        context.CancelFunc                        // 取消函数
	mu            sync.Mutex                                // 互斥锁
	log           *zap.SugaredLogger                        // 日志接口
	connState     *fmsg.ClientConnectState                  //连接状态
	outUMsgChan   chan *fmsg.UMsg                           //需要推送的消息队列
	recUMsgFunMap map[*fmsg.RecEventFunc]*fmsg.RecEventFunc //接收消息回调

	cfg *Option        // 配置
	wsc *utwsc2.Client // Ws客户端接口

	timeDuration time.Duration // 定时器时长
	ticker       *time.Ticker  // 定时器
	TimedEvent   *func()       // 定时器事件
	isTrans      bool
}

func (u *UWsc) SubscribeTimerEvent(fun *func()) {
	u.TimedEvent = fun
}

func (u *UWsc) UnSubscribeTimerEvent(fun *func()) {
	u.TimedEvent = nil
}

func (u *UWsc) GetConnectState() fmsg.ClientConnectState {
	return *u.connState
}

func (u *UWsc) Publish(msg *fmsg.UMsg) {
	u.outUMsgChan <- msg
}

func (u *UWsc) SubscribeRecEvent(fun *fmsg.RecEventFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.recUMsgFunMap[fun] = fun
}

func (u *UWsc) UnSubscribeRecEvent(fun *fmsg.RecEventFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.recUMsgFunMap, fun)
}

func (u *UWsc) Start(ctx context.Context) (done <-chan struct{}, err error) {
	lCtx, cancelFunc := context.WithCancel(ctx)
	u.ctx = ctx
	u.cancel = cancelFunc
	doneC := make(chan struct{})
	cfg1 := utwsc2.Option{
		RemoteIP1:              u.cfg.RemoteIP1,
		RemoteIP2:              u.cfg.RemoteIP2,
		RemotePort:             u.cfg.RemotePort,
		RemotePath:             u.cfg.RemotePath,
		WriteWaitSecond:        u.cfg.WriteWaitSecond,
		HandshakeTimeoutSecond: u.cfg.HandshakeTimeoutSecond,
		RedialIntervalSecond:   u.cfg.RedialIntervalSecond,
		IsPing:                 u.cfg.IsPing,
		PongWaitSecond:         u.cfg.PongWaitSecond,
	}
	d1, err := u.wsc.StartService2(ctx, cfg1, u.Name, u.isTrans)
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(doneC)
		defer func() {
			if err := recover(); err != nil {
				// 记录 panic 信息
				u.log.Errorf("Ws客户端[%s],服务异常,中止服务,错误: %s", u.Name, err)
				// 打印堆栈跟踪信息（可选）
				u.log.Errorf(utils.StackSkip(1, -1))
			}
		}()
		if u.TimedEvent != nil && u.timeDuration > 0 {
			u.ticker = time.NewTicker(u.timeDuration)
		} else {
			u.ticker = time.NewTicker(1)
			u.ticker.Stop()
		}
		defer u.ticker.Stop()
		for {
			select {
			case <-lCtx.Done():
				{
					u.log.Infof("Ws客户端[%s],请求服务停止", u.Name)
					return
				}
			case <-d1:
				{
					u.log.Warnf("Ws客户端[%s]->wsc,服务中止", u.Name)
					return
				}
			case <-u.ticker.C:
				{
					if u.TimedEvent != nil {
						(*u.TimedEvent)()
					}
				}
			case msg, ok := <-u.outUMsgChan:
				{
					if ok {
						if !u.connState.IsConnected {
							u.log.Warnf("Ws客户端[%s],推送消息失败,原因: %s", u.Name, "连接中断")
							break
						}
						u.log.Debugf("Ws客户端[%s],推送消息: %s", u.Name, msg.String("推送"))
						err1 := u.wsc.SendMessage(*msg.Msg)
						if err1 != nil {
							u.log.Errorf("Ws客户端[%s],推送消息:%s,失败: %s,尝试重新连接……", u.Name, msg.String("推送失败"), err1.Error())
							_, err := u.wsc.RestartService()
							if err != nil {
								u.log.Errorf("Ws客户端[%s],尝试重新连接,发送异常,错误: %s", u.Name, err.Error())
							}
						}
					}
				}
			case msg, ok := <-u.wsc.ReceiveMsg:
				{
					if ok {
						if u.recUMsgFunMap != nil && len(u.recUMsgFunMap) > 0 {
							var msgBody msgm.TAiAgentMessage
							err1 := json.Unmarshal(msg.MessageData, &msgBody)
							if err1 != nil {
								u.log.Errorf("Ws客户端[%s],解析消息失败,错误: %s", u.Name, err1.Error())
							}
							outMsg := fmsg.NewUMsg(&msgBody, u.Name, nil)
							u.log.Debugf("Ws客户端[%s],收到消息: %s", u.Name, outMsg.String("接收"))
							u.mu.Lock()
							for _, fun := range u.recUMsgFunMap {
								if fun != nil {
									(*fun)(outMsg)
								}
							}
							u.mu.Unlock()
						}
					}
				}
			case msg, ok := <-u.wsc.ClientConnState:
				{
					if ok {
						stateDes := "连接中断"
						if msg.IsConnected {
							stateDes = "连接成功"
						}
						u.log.Infof("Ws客户端[%s],与Ws服务端: %s:%d,%s", u.Name, u.cfg.RemoteIP1, u.cfg.RemotePort, stateDes)
						if u.recUMsgFunMap != nil && len(u.recUMsgFunMap) > 0 {
							stateDesc := ""
							//生成消息：连接到数据源服务状态发生变化
							var oldState socsws.LinkToDataSourceState
							var newState socsws.LinkToDataSourceState

							if u.connState.IsConnected {
								oldState = socsws.LDS_State_Connected
							} else {
								oldState = socsws.LDS_State_Disconnected
							}
							if msg.IsConnected {
								newState = socsws.LDS_State_Connected
							} else {
								newState = socsws.LDS_State_Disconnected
							}
							if msg.IsConnected {
								stateDesc = "成功连接:" + u.Name
							} else {
								stateDesc = "断开连接:" + u.Name
							}
							*u.connState = msg
							outlsm := socsws.NewMessageForComLinkToDataSourceStateChanged("", oldState, newState, socsws.DataSourceID(u.Name), stateDesc)

							outMsg := fmsg.NewUMsg(&outlsm, u.Name, []string{fmsg.ToWsServer})
							u.mu.Lock()
							for _, fun := range u.recUMsgFunMap {
								if fun != nil {
									(*fun)(outMsg)
								}
							}
							u.mu.Unlock()
						}
					}
				}
			}
		}
	}()
	return doneC, nil
}

func (u *UWsc) Stop() error {
	if u.cancel == nil {
		return nil
	}
	u.cancel()
	err := u.wsc.StopService()
	if err != nil {
		return err
	}
	u.cancel = nil
	return nil
}

func (u *UWsc) RestStart() (done <-chan struct{}, err error) {
	err = u.Stop()
	if err != nil {
		return nil, err
	}
	return u.Start(u.ctx)
}

var _ If = &UWsc{}
