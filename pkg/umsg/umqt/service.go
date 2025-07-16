package umqt

import (
	"context"
	"encoding/json"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/dalm/socsm/socsws"
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
	"github.com/freedqo/fmc-go-agent/pkg/utils"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Option struct {
	Enable   bool     `comment:"是否启用"`    //是否启用
	Broker   []string `comment:"Broker 地址"` // Broker 地址
	User     string   `comment:"用户名"`      // 用户名
	Password string   `comment:"密码"`        // 密码
	SubTopic []string `comment:"订阅的主题"`  // 订阅的主题
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
func New(name string, opt *Option, log *zap.SugaredLogger) umsg.MessageAgentIf {
	s := UMqtt{
		Name: name,
		mu:   sync.Mutex{},
		opt:  opt,
		connState: &umsg.ClientConnectState{
			ClientID:    name,
			IsConnected: false,
		},
		outUMsgChan:   make(chan *umsg.UMsg, 1024*1),
		inUMsgChan:    make(chan *umsg.UMsg, 1024*1),
		recUMsgFunMap: make(map[*umsg.RecEventFunc]*umsg.RecEventFunc),
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
	connState     *umsg.ClientConnectState                  //连接状态
	outUMsgChan   chan *umsg.UMsg                           //需要推送的消息队列
	inUMsgChan    chan *umsg.UMsg                           //需要推送的消息队列
	recUMsgFunMap map[*umsg.RecEventFunc]*umsg.RecEventFunc //接收消息回调

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
						if token := (*u.mq).Publish(msg.Msg.Operate, 0, msg.Msg.Retained, marshal); token.Wait() && token.Error() != nil {
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
// 入参： msg *umsg.UMsg 消息
// 返回： 无
func (u *UMqtt) Publish(msg *umsg.UMsg) {
	u.outUMsgChan <- msg
}

// SubscribeRecEvent 订阅消息
// 入参： fun *umsg.RecEventFunc 回调函数
// 返回： 无
func (u *UMqtt) SubscribeRecEvent(fun *umsg.RecEventFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.recUMsgFunMap[fun] = fun
}

// UnSubscribeRecEvent 取消订阅消息
// 入参： fun *umsg.RecEventFunc 回调函数
// 返回： 无
func (u *UMqtt) UnSubscribeRecEvent(fun *umsg.RecEventFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.recUMsgFunMap, fun)
}

var _ If = &UMqtt{}

// GetConnectState 获取连接状态
// 入参： 无
// 返回： utwebsocket.ClientConnectState 连接状态
func (u *UMqtt) GetConnectState() umsg.ClientConnectState {
	return *u.connState
}

// onReceiveMsg 处理接收到的消息
// 入参： msg *umsg.UMsg
// 返回： 无
func (u *UMqtt) onReceiveUMsg(msg *umsg.UMsg) {
	u.mu.Lock()
	defer u.mu.Unlock()
	msg.Msg.ClientID = u.Name
	msg.Flag = u.Name
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
	stateDesc := ""
	//生成消息：连接到数据源服务状态发生变化
	var oldState socsws.LinkToDataSourceState
	var newState socsws.LinkToDataSourceState
	msg := umsg.ClientConnectState{
		ClientID:    u.Name,
		IsConnected: true,
	}

	if u.connState.IsConnected {
		oldState = socsws.LDS_State_Connected
	} else {
		oldState = socsws.LDS_State_Disconnected
	}

	newState = socsws.LDS_State_Connected

	if msg.IsConnected {
		stateDesc = "成功连接:" + u.Name
	} else {
		stateDesc = "断开连接:" + u.Name
	}
	*u.connState = msg
	outlsm := socsws.NewMessageForComLinkToDataSourceStateChanged("", oldState, newState, socsws.DataSourceID(u.Name), stateDesc)

	outMsg := umsg.NewUMsg(&outlsm, u.Name, []string{umsg.ToWsServer})
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
	stateDesc := ""
	//生成消息：连接到数据源服务状态发生变化
	var oldState socsws.LinkToDataSourceState
	var newState socsws.LinkToDataSourceState
	msg := umsg.ClientConnectState{
		ClientID:    u.Name,
		IsConnected: false,
	}
	if u.connState.IsConnected {
		oldState = socsws.LDS_State_Connected
	} else {
		oldState = socsws.LDS_State_Disconnected
	}

	newState = socsws.LDS_State_Disconnected

	if msg.IsConnected {
		stateDesc = "成功连接:" + u.Name
	} else {
		stateDesc = "断开连接:" + u.Name
	}
	*u.connState = msg
	outlsm := socsws.NewMessageForComLinkToDataSourceStateChanged("", oldState, newState, socsws.DataSourceID(u.Name), stateDesc)
	outMsg := umsg.NewUMsg(&outlsm, u.Name, []string{umsg.ToWsServer})
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
	utMsg := &umsg.Message{}
	err := json.Unmarshal(message.Payload(), utMsg)
	if err != nil {
		message.Topic()
		u.log.Errorf("MQTT[%s],消息解析失败:%s,id:%d,topic:%s,body:%s", u.Name, err.Error(), message.MessageID(), message.Topic(), string(message.Payload()))
		return
	}
	uMsg := umsg.NewUMsg(utMsg, u.Name, nil) // 抛到消息总线，让消息处理处理
	u.inUMsgChan <- uMsg
}
