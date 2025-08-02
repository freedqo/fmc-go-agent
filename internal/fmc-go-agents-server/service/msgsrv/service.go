package msgsrv

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/msgm"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/msgsrv/handler"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/msgsrv/hub"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/store/log"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	"sync"
	"time"
)

const InMsgCacheSize = 1024 * 10
const OutMsgCacheSize = 1024 * 10

// New 创建消息总线服务
// 入参： 配置文件
// 返回： 消息总线服务接口实例
func New(cfg *config.Config) If {
	s := Service{
		cfg:             cfg,
		mu:              sync.Mutex{},
		hub:             hub.New(cfg),
		inMsgChan:       make(chan *fmsg.UMsg, InMsgCacheSize),  //需要处理的消息队列
		outMsgChan:      make(chan *fmsg.UMsg, OutMsgCacheSize), //需要推送的消息队列
		handlerInMsgFun: make(map[*fmsg.HandlerInMsgFun]*fmsg.HandlerInMsgFun),
	}
	// 注册消息处理服务,消息输入方法
	s.vPublish = s.Publish

	return &s
}

type Service struct {
	cfg    *config.Config
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	hub    hub.If

	inMsgChan  chan *fmsg.UMsg //需要处理的消息队列
	outMsgChan chan *fmsg.UMsg //需要推送的消息队列

	handlerInMsgFun   map[*fmsg.HandlerInMsgFun]*fmsg.HandlerInMsgFun //待处理消息消息的处理器
	vHandlerInMsgFunc fmsg.HandlerInMsgFun                            //固化的待处理消息的处理器

	vPublish fmsg.RecEventFunc
}

// GetMsgHubIf 获取消息hub接口
// 入参： name string 消息hub名称
// 返回： 消息hub接口实例
func (s *Service) GetMsgHubIf(name string) fmsg.MessageHubIf {
	return s.hub.GetMsgHubIf(name)
}

// Publish 发布
// 入参： msg *fmsg.UMsg 消息
// 返回： 无
func (s *Service) Publish(msg *fmsg.UMsg) {
	if len(s.inMsgChan) > (InMsgCacheSize - 100) {
		log.SysLog().Warnf("输入消息队列,当前长度:%d,缓存接近饱和,存在堵塞风险", len(s.inMsgChan))
	}
	log.SysLog().Infof("收到消息")
	s.inMsgChan <- msg
}

// WaitingResMsg 函数用于等待应答消息
func (s *Service) WaitingResMsg(operateID string, waitTime time.Duration) (*fmsg.UMsg, error) {
	if waitTime <= 0 {
		return nil, errors.New("等待时间必须大于0")
	}
	if operateID == "" {
		return nil, errors.New("操作ID不能为空")
	}
	// 创建一个带有超时的上下文
	ctx, cancel := context.WithTimeout(s.ctx, waitTime)
	defer cancel()

	// 创建一个消息通道，用于接收消息
	msgChan := make(chan *fmsg.UMsg, 1)
	// 定义一个消息处理函数
	var vhandler fmsg.HandlerInMsgFun
	// 是否已经接收到了消息
	isRec := false
	// 定义消息处理函数
	vhandler = func(push fmsg.PublishFunc, msg *fmsg.UMsg) {
		// 处理消息结构体
		buf, ok := msg.Msg.([]byte)
		if !ok {
			return
		}
		var d msgm.TAiAgentMessage
		err := json.Unmarshal(buf, &d)
		if err != nil {
			return
		}
		// 如果还没有接收到消息,且消息的OperateID与传入的operateID相等,且消息是应答消息
		if !isRec && d.OperateID == operateID && d.RespType == msgm.RespType_ClientRespServer {
			// 将消息发送到消息通道
			msgChan <- msg
			// 设置已经接收到消息
			isRec = true
		}
	}
	// 定义一个变量用于存储接收到的消息
	var resMsg *fmsg.UMsg
	// 注册监听
	s.Subscribe(&vhandler)
	defer s.UnSubscribe(&vhandler)

	// 使用select语句等待消息
	select {
	// 如果上下文超时
	case <-ctx.Done():
		{
			// 记录日志
			log.SysLog().Errorf("等待应答消息超时")
			break
		}
	// 如果从消息通道中接收到消息
	case msg, ok := <-msgChan:
		{
			// 如果消息通道没有关闭
			if ok {
				// 将接收到的消息赋值给resMsg
				resMsg = msg
				log.SysLog().Errorf("收到需要的消息了")
				break
			}
		}
	}
	// 如果resMsg为空
	if resMsg == nil {
		// 返回错误
		return nil, ctx.Err()
	} else {
		// 返回接收到的消息
		return resMsg, nil
	}
}

// push 推送
// 入参： msg *fmsg.UMsg 消息
// 返回： 无
func (s *Service) push(msg *fmsg.UMsg) {
	if len(s.outMsgChan) > (InMsgCacheSize - 100) {
		log.SysLog().Warnf("输出消息队列,当前长度:%d,缓存接近饱和,存在堵塞风险", len(s.outMsgChan))
	}
	s.outMsgChan <- msg
}

// Subscribe 订阅
// 入参： fun *fmsg.HandlerInMsgFun 消息处理器
// 返回： 无
func (s *Service) Subscribe(fun *fmsg.HandlerInMsgFun) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlerInMsgFun[fun] = fun
}

// UnSubscribe 取消订阅
// 入参： fun *fmsg.HandlerInMsgFun 消息处理器
// 返回： 无
func (s *Service) UnSubscribe(fun *fmsg.HandlerInMsgFun) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.handlerInMsgFun, fun)
}

// Start 启动服务
// 入参： ctx context.Context 上下文
// 返回： done <-chan struct{} 服务结束信号
// 返回： err error 错误信息
func (s *Service) Start(ctx context.Context) (done <-chan struct{}, err error) {
	s.ctx = ctx
	lctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	doneC := make(chan struct{})
	//注册消息回调
	s.hub.SubscribeRecEvent(&s.vPublish)
	//启动服务
	d1, err := s.hub.Start(lctx)
	var once sync.Once

	// 启动消息处理器
	h := handler.New()
	s.vHandlerInMsgFunc = h.HandleInMsgFunc
	s.Subscribe(&s.vHandlerInMsgFunc) //启动消息处理

	go s.handleMessages(lctx, doneC, &once)
	//启动消息推送
	go s.pushMessages(lctx, doneC, &once, d1)
	return doneC, nil
}

// Stop 停止服务
// 入参： 无
// 返回： err error 错误信息
func (s *Service) Stop() error {
	if s.cancel == nil {
		return nil
	}
	// 注销消息回调
	s.hub.UnSubscribeRecEvent(&s.vPublish)
	s.cancel()
	s.cancel = nil
	return nil
}

// RestStart 重启服务
// 入参： 无
// 返回： done <-chan struct{} 服务结束信号
func (s *Service) RestStart() (done <-chan struct{}, err error) {
	err = s.Stop()
	if err != nil {
		return nil, err
	}
	return s.Start(s.ctx)
}

// handleMessages 处理输入消息队列中的消息
// 入参： ctx context.Context 上下文
// 入参： doneC chan struct{} 服务结束信号
// 入参： once *sync.Once 同步锁
// 返回： 无
func (s *Service) handleMessages(ctx context.Context, doneC chan struct{}, once *sync.Once) {
	defer func() {
		once.Do(func() {
			close(doneC)
		})
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-s.inMsgChan:
			if ok {
				s.mu.Lock()
				for _, fun := range s.handlerInMsgFun {
					if fun != nil {
						(*fun)(s.push, msg)
					}
				}
				s.mu.Unlock()
			}
		}
	}
}

// pushMessages 推送输出消息队列中的消息
// 入参： ctx context.Context 上下文
// 入参： doneC chan struct{} 服务结束信号
// 入参： once *sync.Once 同步锁
// 入参： d1 <-chan struct{} 消息hub结束信号
// 返回： 无
func (s *Service) pushMessages(ctx context.Context, doneC chan struct{}, once *sync.Once, d1 <-chan struct{}) {
	defer func() {
		once.Do(func() {
			close(doneC)
		})
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-d1:
			return
		case msg, ok := <-s.outMsgChan:
			if ok {
				s.hub.Publish(msg)
			}
		}
	}
}

var _ If = &Service{}
