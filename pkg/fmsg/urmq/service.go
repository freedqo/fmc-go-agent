package urmq

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/socsm/socsws"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/msgm"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	"github.com/freedqo/fmc-go-agents/pkg/rocketmqx"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
	"sync"
	"time"
)

type Option struct {
	Enable              bool                   `comment:"是否启用"`       // 是否启用
	NameSrvs            []string               `comment:"注册中心"`       // 这个地址相当于注册中心, 会有心跳检测检查集群中的broker, 然后更新进来
	Brokers             []string               `comment:"broker地址列表"` // 这个地址只在创建主题的时候需要
	Credentials         *primitive.Credentials `comment:"Token配置"`
	ConsumerGroupPrefix string                 `comment:"消费者分组前缀"` // 消费者分组前缀
	LogLevel            int                    `comment:"日志等级"`
	// 该配置是为确保**不同**服务中订阅了相同主题的消费者分在不同的群组(群组名不重复)
	// 从而不同服务可以各自收到消息并独立地处理消息
	// 群组的作用是使消费者在集群中起到负载均衡的作用, 对于**集群模式**的消息, 相同群组的消费者,只要有一个消费了,同群组的消费者就不会再消费了
	// 如果不同服务的消费者有相同的群组, 就很可能会发生一条消息只被其中一个服务的消费者消费
}

func NewDefaultOption() *Option {
	return &Option{
		Enable:      false,
		NameSrvs:    []string{"127.0.0.1:9876"},
		Brokers:     []string{"127.0.0.1:10911"},
		Credentials: nil,
	}
}

// New 创建客户端实例
func New(name string, o *Option, topicList []string, log *zap.SugaredLogger) *URocketMq {
	if len(o.ConsumerGroupPrefix) == 0 {
		panic("消费者群组名称前缀不能为空， 请检查RocketMQ配置项进行配置")
	}
	r := &URocketMq{
		Name:        name,
		log:         log,
		mu:          sync.Mutex{},
		outUMsgChan: make(chan *fmsg.UMsg, 1024*1),
		connState: &fmsg.ClientConnectState{
			ClientID:    name,
			IsConnected: false,
		},
		recUMsgFunMap: map[*fmsg.RecEventFunc]*fmsg.RecEventFunc{},
		TopicList:     topicList,
		cfg:           o,
	}
	return r
}

type URocketMq struct {
	Name          string                                    // 客户端名称
	ctx           context.Context                           // 上下文
	cancel        context.CancelFunc                        // 取消函数
	mu            sync.Mutex                                // 互斥锁
	log           *zap.SugaredLogger                        // 日志接口
	connState     *fmsg.ClientConnectState                  //连接状态
	outUMsgChan   chan *fmsg.UMsg                           //需要推送的消息队列
	recUMsgFunMap map[*fmsg.RecEventFunc]*fmsg.RecEventFunc //接收消息回调

	mq        *rocketmqx.RocketMQ // RocketMq客户端接口
	cfg       *Option             // 配置
	TopicList []string            // 订阅的主题
}

func (u *URocketMq) GetConnectState() fmsg.ClientConnectState {
	//TODO implement me
	panic("implement me")
}

func (u *URocketMq) Publish(msg *fmsg.UMsg) {
	u.outUMsgChan <- msg
}

func (u *URocketMq) SubscribeRecEvent(fun *fmsg.RecEventFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.recUMsgFunMap[fun] = fun
}

func (u *URocketMq) UnSubscribeRecEvent(fun *fmsg.RecEventFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.recUMsgFunMap, fun)
}

// Start 启动
func (u *URocketMq) Start(ctx context.Context) (done <-chan struct{}, err error) {
	lctx, cancel := context.WithCancel(ctx)
	u.ctx = ctx
	u.cancel = cancel
	d := make(chan struct{})
	//实例化rocketmq
	u.mq = rocketmqx.NewRocketMQ(u.Name, &rocketmqx.RocketMQOptions{
		NameSrvs:            u.cfg.NameSrvs,
		Brokers:             u.cfg.Brokers,
		Credentials:         u.cfg.Credentials,
		LogLevel:            zapcore.Level(u.cfg.LogLevel),
		ConsumerGroupPrefix: u.cfg.ConsumerGroupPrefix,
	}, u.log)
	// 注册主题
	err = u.mq.BatchCreateTopic(lctx, u.TopicList, u.mq.NameSrvs, u.mq.Brokers)
	if err != nil {
		return nil, err
	}
	// 等待创建主题完成
	time.Sleep(time.Millisecond * 10)
	// 注册消费者
	for _, topic := range u.TopicList {
		err = u.mq.Consumer(lctx, topic, u.handleReceiveMsg)
		if err != nil {
			return nil, err
		}
	}
	// 启动消费者
	for _, pushConsumer := range u.mq.Consumers {
		err = pushConsumer.Start()
		if err != nil {
			u.log.Errorf("RocketMq[%s] 消费者,启动失败: %s", u.Name, err.Error())
			return nil, err
		}
	}
	u.handleOnConnOnLine()
	go func() {
		defer func() {
			for _, p := range u.mq.Producers {
				err := p.Shutdown()
				if err != nil {
					u.log.Errorf("RocketMq[%s] 生产者,关闭失败: %s", u.Name, err.Error())
				} else {
					u.log.Debugf("RocketMq[%s] 生产者,关闭成功", u.Name)
				}
			}
			for _, c := range u.mq.Consumers {
				err := c.Shutdown()
				if err != nil {
					u.log.Errorf("RocketMq[%s] 消费者,关闭失败: %s", u.Name, err.Error())
				} else {
					u.log.Debugf("RocketMq[%s] 消费者,关闭成功", u.Name)
				}
			}
			close(d)
		}()
		t := time.NewTicker(time.Second * 60)
		defer t.Stop()
		for {
			select {
			case <-lctx.Done():
				return
			case <-t.C:
				{
					u.handleHeartBeat()
				}
			case msg, ok := <-u.outUMsgChan:
				{
					if ok {
						omsg := rocketmqx.NewMessage(msg.Msg.Topic)
						omsg.AddBody(msg.Msg.Data)
						if msg.IsASync {
							err := u.mq.PublishAsync(lctx, omsg, nil)
							if err != nil {
								u.log.Errorf("RocketMq[%s],推送消息:%s,失败: %s", u.Name, msg.String("推送失败"), err.Error())
							}
						} else {
							err := u.mq.PublishSync(lctx, omsg)
							if err != nil {
								u.log.Errorf("RocketMq[%s],推送消息:%s,失败: %s", u.Name, msg.String("推送失败"), err.Error())
							}
						}
					}
				}
			}
		}
	}()
	return d, nil
}

// Stop 停止
func (u *URocketMq) Stop() error {
	if u.cancel == nil {
		return nil
	}
	u.cancel()
	u.cancel = nil
	for _, p := range u.mq.Producers {
		err := p.Shutdown()
		if err != nil {
			u.log.Errorf("RocketMq[%s] 生产者,关闭失败: %s", u.Name, err.Error())
		} else {
			u.log.Debugf("RocketMq[%s] 生产者,关闭成功", u.Name)
		}
	}
	for _, c := range u.mq.Consumers {
		err := c.Shutdown()
		if err != nil {
			u.log.Errorf("RocketMq[%s] 消费者,关闭失败: %s", u.Name, err.Error())
		} else {
			u.log.Debugf("RocketMq[%s] 消费者,关闭成功", u.Name)
		}
	}
	u.handleOnConnOffLine()

	return nil
}

// RestStart 重启
func (u *URocketMq) RestStart() (done <-chan struct{}, err error) {
	err = u.Stop()
	if err != nil {
		return nil, err
	} // 先关闭
	return u.Start(u.ctx)
}

var _ If = &URocketMq{}

func (u *URocketMq) onReceiveUMsg(msg *fmsg.UMsg) {
	u.mu.Lock()
	defer u.mu.Unlock()
	msg.Msg.SessionId = u.Name
	msg.Ctx = u.Name
	for _, item := range u.recUMsgFunMap {
		(*item)(msg)
	}
}

// handleReceiveMsg RocketMq消息接收转换器
func (u *URocketMq) handleReceiveMsg(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	defer func() {
		if err := recover(); err != nil {
			// 生成一个唯一的错误ID，用于后续的错误跟踪
			PanicId := uuid.New().ID()
			// 记录 panic 信息
			u.log.Errorf("RocketMq[%s],处理待接收消息,遇到异常,PanicId: %d, Panic: %v", u.Name, PanicId, err)
			// 打印堆栈跟踪信息（可选）
			u.log.Errorf(utils.StackSkip(1, -1))
		}
	}()
	for _, msg := range msgs {
		if msg == nil || msg.Body == nil || strings.TrimSpace(msg.Topic) == "" {
			continue
		}
		umsg := fmsg.NewUMsg(&msgm.TAiAgentMessage{
			MessageBase: fmsg.MessageBase{
				SessionId:       u.Name,
				Topic:           msg.Topic,
				RespType:        false,
				OperateID:       msg.MsgId,
				OperateDataType: fmt.Sprintf("%T", msg),
			},
			Data: msg,
		}, u.Name, nil)
		u.log.Debugf("RocketMq[%s],收到消息:%s", u.Name, umsg.String("接收"))
		u.onReceiveUMsg(umsg)
	}

	return consumer.ConsumeSuccess, nil
}

func (u *URocketMq) handleOnConnOnLine() {
	if u.connState.IsConnected {
		return
	}
	stateDesc := ""
	//生成消息：连接到数据源服务状态发生变化
	var oldState socsws.LinkToDataSourceState
	var newState socsws.LinkToDataSourceState
	msg := fmsg.ClientConnectState{
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

	outMsg := fmsg.NewUMsg(&outlsm, u.Name, []string{fmsg.ToWsServer})
	u.onReceiveUMsg(outMsg)
	u.log.Infof("RocketMq[%s],连接成功", u.Name)
}

func (u *URocketMq) handleOnConnOffLine() {
	if !u.connState.IsConnected {
		return
	}
	stateDesc := ""
	//生成消息：连接到数据源服务状态发生变化
	var oldState socsws.LinkToDataSourceState
	var newState socsws.LinkToDataSourceState
	msg := fmsg.ClientConnectState{
		ClientID:    u.Name,
		IsConnected: false,
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

	outMsg := fmsg.NewUMsg(&outlsm, u.Name, []string{fmsg.ToWsServer})
	u.onReceiveUMsg(outMsg)
	u.log.Infof("RocketMq[%s],连接中断", u.Name)
}

func (u *URocketMq) handleHeartBeat() {
	testAdmin, err := admin.NewAdmin(admin.WithResolver(primitive.NewPassthroughResolver(u.cfg.NameSrvs)))
	if err != nil {
		u.log.Errorf(fmt.Sprintf("RocketMq[%s],连接失败: %s", u.Name, err))
		u.handleOnConnOffLine()
		return
	}
	defer testAdmin.Close()
	u.handleOnConnOnLine()
	topicList, err := testAdmin.FetchAllTopicList(u.ctx)
	if err != nil {
		u.log.Errorf(fmt.Sprintf("RocketMq[%s],连接失败: %s", u.Name, err))
		u.handleOnConnOffLine()
		return
	}
	for _, topicName := range u.TopicList {
		isExit := false
		for _, item := range topicList.TopicList {
			if strings.Contains(item, topicName) {
				isExit = true
				break
			}
		}
		if !isExit {
			for _, broker := range u.mq.Brokers {
				err = testAdmin.CreateTopic(u.ctx, admin.WithTopicCreate(topicName), admin.WithBrokerAddrCreate(broker))
				if err != nil {
					u.log.Errorf("RocketMq[%s],broker:%s 创建主题 %s 失败: %s", u.Name, broker, topicName, err.Error())
				} else {
					u.log.Debugf("RocketMq[%s],broker:%s 创建主题: %s 成功", u.Name, broker, topicName)
				}
			}
		}
	}
}
