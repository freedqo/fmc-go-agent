package umqt

import (
	"context"
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Option struct {
	Enable   bool     `comment:"是否启用"`      //是否启用
	Broker   []string `comment:"Broker 地址"` // Broker 地址
	User     string   `comment:"用户名"`       // 用户名
	Password string   `comment:"密码"`        // 密码
	SubTopic []string `comment:"订阅的主题"`     // 订阅的主题
}

func NewDefaultOption() *Option {
	return &Option{
		Enable:   false,
		Broker:   []string{"tcp://localhost:1883"},
		User:     "admin",
		Password: "Unitech@1998",
		SubTopic: []string{""},
	}
}

// New 创建一个 MQTT 客户端实例
// 入参： name string 客户端名称
// 入参： opt *config.MqttLinkCfg 配置
// 入参： topicList  []string 订阅的主题列表
// 入参： log *zap.SugaredLogger 日志接口
// 返回： *UMqtt 客户端实例
func New(name string, opt *Option, log *zap.SugaredLogger) fmsg.MessageAgentIf {
	s := UMqtt{
		Name: name,
		mu:   sync.Mutex{},
		opt:  opt,
		connState: &fmsg.TConnStatus{
			Name:  name,
			State: false,
		},
		outUMsgChan:   make(chan *fmsg.UMsg, 1024*1),
		inUMsgChan:    make(chan *fmsg.UMsg, 1024*1),
		recUMsgFunMap: make(map[*fmsg.RecEventFunc]*fmsg.RecEventFunc),
		log:           log,
	}

	return &s

}

type UMqtt struct {
	Name          string                                    // 客户端名称
	ctx           context.Context                           // 上下文
	cancel        context.CancelFunc                        // 取消函数
	mu            sync.Mutex                                // 互斥锁
	log           *zap.SugaredLogger                        // 日志接口
	connState     *fmsg.TConnStatus                         //连接状态
	outUMsgChan   chan *fmsg.UMsg                           //需要推送的消息队列
	inUMsgChan    chan *fmsg.UMsg                           //需要推送的消息队列
	recUMsgFunMap map[*fmsg.RecEventFunc]*fmsg.RecEventFunc //接收消息回调

	mq        *mqtt.Client        // Mqtt客户端接口
	opt       *Option             // 配置
	utMqttOpt *mqtt.ClientOptions // Mqtt客户端选项
}

// Start 启动 MQTT 客户端
// 入参： ctx context.Context 上下文
// 返回： done <-chan struct{} 完成通道
// 返回： err error 错误信息
func (u *UMqtt) Start(ctx context.Context) (done <-chan struct{}, err error) {
	lCtx, cancelFunc := context.WithCancel(ctx)
	u.ctx = ctx
	u.cancel = cancelFunc
	doneC := make(chan struct{})
	// 创建 MQTT 客户端选项
	opts := mqtt.NewClientOptions()
	for _, broker := range u.opt.Broker {
		opts.AddBroker(broker)
	}
	opts.SetClientID(u.Name)
	opts.SetUsername(u.opt.User)
	opts.SetPassword(u.opt.Password)
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetMaxReconnectInterval(time.Second * 30)
	opts.SetConnectRetry(true)
	opts.SetDefaultPublishHandler(u.handleReceiveMsg)
	opts.OnConnectionLost = u.handlerOnConnectionLost
	opts.OnReconnecting = u.handlerOnReconnecting
	opts.OnConnect = u.handlerOnConnect
	client := mqtt.NewClient(opts)

	u.utMqttOpt = opts
	u.mq = &client
	// 连接到 MQTT 服务器
	if token := (*u.mq).Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	go func() {
		defer func() {
			(*u.mq).Disconnect(250)
			close(doneC)
		}()
		for {
			select {
			case <-lCtx.Done():
				{
					return
				}
			case msg, ok := <-u.outUMsgChan:
				{
					if ok {
						u.log.Debugf("MQTT[%s],推送消息: %s", u.Name, msg.String("推送"))
						marshal, err := json.Marshal(msg.Msg)
						if err != nil {
							u.log.Errorf("MQTT[%s],推送消息:%s,失败,错误:%s", u.Name, msg.String("推送失败"), err.Error())
							return
						}
						if token := (*u.mq).Publish(msg.Topic, 0, false, marshal); token.Wait() && token.Error() != nil {
							u.log.Errorf("MQTT[%s],推送消息:%s,失败,错误:%s", u.Name, msg.String("推送失败"), token.Error())
						}
					}
				}
			case msg, ok := <-u.inUMsgChan:
				{
					if ok {
						u.log.Debugf("MQTT[%s],收到消息:%s", u.Name, msg.String("接收"))
						u.onReceiveUMsg(msg)
					}
				}
			}
		}
	}()
	return doneC, nil
}

// Stop 停止 MQTT 客户端
// 入参： 无
// 返回： err error 错误信息
func (u *UMqtt) Stop() error {
	if u.cancel == nil {
		return nil
	}
	(*u.mq).Disconnect(250)
	u.cancel()
	u.cancel = nil
	return nil
}

// RestStart 重启 MQTT 客户端
// 入参： 无
// 返回： done <-chan struct{} 完成通道
// 返回： err error 错误信息
func (u *UMqtt) RestStart() (done <-chan struct{}, err error) {
	err = u.Stop()
	if err != nil {
		return nil, err
	}
	return u.Start(u.ctx)
}

// Publish 推送消息
// 入参： msg *fmsg.UMsg 消息
// 返回： 无
func (u *UMqtt) Publish(msg *fmsg.UMsg) {
	u.outUMsgChan <- msg
}

// SubscribeRecEvent 订阅消息
// 入参： fun *fmsg.RecEventFunc 回调函数
// 返回： 无
func (u *UMqtt) SubscribeRecEvent(fun *fmsg.RecEventFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.recUMsgFunMap[fun] = fun
}

// UnSubscribeRecEvent 取消订阅消息
// 入参： fun *fmsg.RecEventFunc 回调函数
// 返回： 无
func (u *UMqtt) UnSubscribeRecEvent(fun *fmsg.RecEventFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.recUMsgFunMap, fun)
}

var _ If = &UMqtt{}

// GetConnectState 获取连接状态
// 入参： 无
// 返回： utwebsocket.ClientConnectState 连接状态
func (u *UMqtt) GetConnectState() fmsg.TConnStatus {
	return *u.connState
}

// onReceiveMsg 处理接收到的消息
// 入参： msg *fmsg.UMsg
// 返回： 无
func (u *UMqtt) onReceiveUMsg(msg *fmsg.UMsg) {
	u.mu.Lock()
	defer u.mu.Unlock()
	msg.Sour = u.Name
	for _, item := range u.recUMsgFunMap {
		(*item)(msg)
	}
}

// handlerOnConnect 处理接收到的消息
// 入参： client mqtt.Client 客户端
// 返回： 无
func (u *UMqtt) handlerOnConnect(client mqtt.Client) {
	for _, topic := range u.opt.SubTopic {
		u.log.Debugf("MQTT[%s] 订阅主题成功:%s", u.Name, topic)
		if token := (*u.mq).Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
			u.log.Errorf("MQTT[%s] 订阅主题失败:%s", u.Name, token.Error())
		}
	}
	outMsg := fmsg.NewUMsg("TConnStatus", &fmsg.TConnStatus{
		Name:  u.Name,
		State: true,
	}, u.Name, nil, nil)
	u.onReceiveUMsg(outMsg)
	u.log.Infof("MQTT[%s],连接成功", u.Name)
}

// handlerOnConnectionLost 处理连接中断
// 入参： client mqtt.Client 客户端
// 返回： 无
func (u *UMqtt) handlerOnConnectionLost(client mqtt.Client, err error) {
	defer func() {
		if err := recover(); err != nil {
			// 生成一个唯一的错误ID，用于后续的错误跟踪
			PanicId := uuid.New().ID()
			// 记录 panic 信息
			u.log.Errorf("MQTT[%s],处理异常中断回调,遇到异常,PanicId: %d, Panic: %v", u.Name, PanicId, err)
			// 打印堆栈跟踪信息（可选）
			u.log.Errorf(utils.StackSkip(1, -1))
		}
	}()
	outMsg := fmsg.NewUMsg("TConnStatus", &fmsg.TConnStatus{
		Name:  u.Name,
		State: false,
	}, u.Name, nil, nil)
	u.onReceiveUMsg(outMsg)
	u.log.Infof("MQTT[%s],连接中断", u.Name)

}

// handlerOnReconnecting 处理重连
// 入参： client mqtt.Client 客户端
// 返回： 无
func (u *UMqtt) handlerOnReconnecting(client mqtt.Client, options *mqtt.ClientOptions) {
	defer func() {
		if err := recover(); err != nil {
			// 生成一个唯一的错误ID，用于后续的错误跟踪
			PanicId := uuid.New().ID()
			// 记录 panic 信息
			u.log.Errorf("RocketMq[%s],处理异常重连成功前再回调,遇到异常,PanicId: %d, Panic: %v", u.Name, PanicId, err)
			// 打印堆栈跟踪信息（可选）
			u.log.Errorf(utils.StackSkip(1, -1))
		}
	}()
}

// handleReceiveMsg 处理接收到的消息
// 入参： client mqtt.Client 客户端
// 入参： message mqtt.Message 消息
// 返回： 无
func (u *UMqtt) handleReceiveMsg(client mqtt.Client, message mqtt.Message) {
	defer func() {
		if err := recover(); err != nil {
			// 生成一个唯一的错误ID，用于后续的错误跟踪
			PanicId := uuid.New().ID()
			// 记录 panic 信息
			u.log.Errorf("MTQQ[%s],处理待接收消息,遇到异常,PanicId: %d, Panic: %v", u.Name, PanicId, err)
			// 打印堆栈跟踪信息（可选）
			u.log.Errorf(utils.StackSkip(1, -1))
		}
	}()
	// 确认消息已被处理
	defer message.Ack()
	uMsg := fmsg.NewUMsg(message.Topic(), message.Payload(), u.Name, nil, nil) // 抛到消息总线，让消息处理处理
	u.inUMsgChan <- uMsg
}
