package utwss

import (
	"context"
	"errors"
	"fmt"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/msgm"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// 构建utwebsocket的服务端
func New(log *zap.SugaredLogger, opt *Option) *Server {
	server := &Server{
		hub:             newHub(log),
		ReceiveMsg:      make(chan fmsg.InMessage, 1024),
		ClientConnState: make(chan fmsg.ClientConnectState, 1024),
		opt:             opt,
		log:             log,
	}
	return server
}

// Server utwebsocket的服务端
type Server struct {
	upgrader        websocket.Upgrader           //websocket服务端
	opt             *Option                      //配置
	hub             *hub                         //客户端管理中心
	ReceiveMsg      chan fmsg.InMessage          //管道：对外输出接收的消息。如果外部不及时取出接收的这些消息，达到最大缓冲后将丢掉旧的消息。
	ClientConnState chan fmsg.ClientConnectState //管道：向外推送客户端的连接状态事件
	cancel          context.CancelFunc           //取消本服务的方法
	parentCtx       context.Context              //调用本服务的context
	log             *zap.SugaredLogger           //专用日志
}

// 启动服务
func (s *Server) StartService(ctx context.Context) (done <-chan struct{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintln(e))
			s.log.Errorf("%v", err) //输出到致命（关键错误）日志跟踪器
		}
	}()
	doneChan := make(chan struct{})
	s.parentCtx = ctx
	myctx, cancel := context.WithCancel(s.parentCtx)
	if s.cancel != nil {
		s.cancel()
		time.Sleep(1 * time.Second)
	}
	s.cancel = cancel
	//websocket路由
	serveMux := http.NewServeMux()
	serveMux.HandleFunc(s.opt.ServicePattern, func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				s.log.Errorf("%v", e) //输出到致命（关键错误）日志跟踪器
			}
		}()
		s.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		//从请求头获取SessionId
		sessionId := r.Header.Get("SessionId")
		if sessionId == "" {
			sessionId = uuid.New().String()
			_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("SessionId为空,禁止连接！")))
			s.log.Errorf("远程客户端[%s],SessionId为空,禁止连接！", conn.RemoteAddr())
			_ = conn.Close()
			return
		}
		_, ok := s.hub.ClientidToClients[sessionId]
		if ok {
			_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("SessionId:%s,已建立连接,禁止重复连接！", sessionId)))
			s.log.Errorf("远程客户端[%s],SessionId:%s,已建立连接,禁止重复连接！", conn.RemoteAddr(), sessionId)
			_ = conn.Close()
			return
		}

		client := newConnClient(sessionId, s.opt, s.hub, conn, s.log)
		var remoteAddr net.Addr = nil
		if client.conn != nil {
			remoteAddr = client.conn.RemoteAddr()
		}
		client.hub.register <- &serverClientWithRemoteAddr{
			ClientID:        sessionId,
			ServerClientObj: client,
			RemoteAddr:      remoteAddr,
		}
		s.log.Debugf("客户端[%s]接入 本地地址：%v 远方地址：%v", sessionId, client.conn.LocalAddr(), client.conn.RemoteAddr())
		client.startService(myctx)
	})
	//侦听服务
	addr := fmt.Sprintf(":%d", s.opt.ServicePort)
	s.log.Debugf("侦听服务：%s", fmt.Sprintf("%s%s", addr, s.opt.ServicePattern))
	go func() {
		defer func() {
			if e := recover(); e != nil {
				s.log.Errorf("%v", e) //输出到致命（关键错误）日志跟踪器
			}
		}()
		defer close(doneChan)

		s.hub.run(myctx)
		s.serviceRecieveMsg(myctx)
		subDone := make(chan struct{})
		go func() {
			defer func() {
				if e := recover(); e != nil {
					s.log.Errorf("%v", e) //输出到致命（关键错误）日志跟踪器
				}
			}()
			defer close(subDone)
			err2 := http.ListenAndServe(addr, serveMux)
			if err2 != nil {
				s.log.Errorf("%s %s %s %v", fmt.Sprintf("%s%s", addr, s.opt.ServicePattern), "ListenAndServe", "错误", err2)
			}
		}()
		select {
		case <-subDone:
		case <-myctx.Done():
		}
		s.log.Debugf("%s 已结束服务", fmt.Sprintf("%s%s", addr, s.opt.ServicePattern))
	}()

	return doneChan, nil
}

// 单独停止本服务
// 也可以通过Context，在统一结束所有服务时，结束本服务。
func (s *Server) StopService() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintln(e))
			s.log.Errorf("%v", err) //输出到致命（关键错误）日志跟踪器
		}
	}()
	s.log.Debugf("请求结束服务:")
	s.cancel()
	return nil
}

// 重启动服务
// 调用RestartService()方法之前，请先结束服务（例如，可以调用StopService()方法）
// 基于上次调用StartService()方法的Context和Cfg启动服务。
// 如果之前没有调用过StartService()方法，则将会返回错误
// 如果Cfg有变化，则先修改Cfg,再调用本方法
// 也可以直接调用StartService()方法重新启动服务。但要注意的是：要正确使用管理本服务Context，以确保外部能够通过统一方式正确结束所有服务（包括本服务）。
func (s *Server) RestartService() (done <-chan struct{}, err error) {

	s.log.Debugf("请求重启服务")
	if s.parentCtx == nil {
		err = errors.New("之前没调用过StartService()方法！")
		s.log.Debugf("请求重启服务错误：%v", err)
		return done, err
	}
	return s.StartService(s.parentCtx)
}

// 发送消息。异步发送，无法确保消息发送到目标客户端。
// msg.ClientID为目标客户端ID;空：表示广播
func (s *Server) SendMessage(msg msgm.TAiAgentMessage) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintln(e))
			s.log.Errorf("%v", err) //输出到致命（关键错误）日志跟踪器
		}
	}()
	s.log.Debugf("请求发送消息 消息内容:%v", msg)
	if len(s.hub.SendMsg) == cap(s.hub.SendMsg) {
		err = errors.New("发送消息队列已满")
		s.log.Debugf("SendMessage()错误：%v", err)
		return err
	}
	if msg.SessionId != "" {
		_, ok := s.hub.ClientidToClients[msg.SessionId]
		if !ok {
			if msg.OperateID != "" {
				msg.MessageBase.RespType = true
				cbMsg, err := fmsg.NewInMessage(msg.MessageBase, "操作发送失败,远程Web客户端不在线")
				if err != nil {
					return fmt.Errorf("目标客户端[%s]离线,且构建回调消息异常：%v", msg.SessionId, err)
				}
				s.ReceiveMsg <- cbMsg
			}
			return fmt.Errorf("目标客户端[%s]离线", msg.SessionId)
		}
	}

	//构建输出消息
	outMsg, err := fmsg.NewOutMessage(msg)
	if err != nil {
		return err
	}
	s.hub.SendMsg <- outMsg

	return nil
}

// 启动接收消息服务
func (s *Server) serviceRecieveMsg(ctx context.Context) (done <-chan struct{}) {
	doneChan := make(chan struct{})
	go func() {
		defer func() {
			if e := recover(); e != nil {
				s.log.Errorf("%v", e) //输出到致命（关键错误）日志跟踪器
			}
		}()
		defer close(doneChan)
		for {
			select {
			case imsg, ok := <-s.hub.ReceiveMsg:
				if ok {
					if len(s.ReceiveMsg) == cap(s.ReceiveMsg) {
						oldMsg := <-s.ReceiveMsg //去掉旧的，保留最新的,以防缓存满后，造成阻塞
						s.log.Debugf("ReceiveMsg消息队列已满，删除最旧的消息：%v", oldMsg)
					}
					s.ReceiveMsg <- imsg
					s.log.Debugf("放入ReceiveMsg队列中的消息：%v", imsg)
				}
			case clientInfo, ok := <-s.hub.UnregisterClientID:
				if ok {
					if len(s.ClientConnState) == cap(s.ClientConnState) {
						oldClientConnState := <-s.ClientConnState //去掉旧的，保留最新的,以防缓存满后，造成阻塞
						s.log.Debugf("ClientConnState消息队列已满，删除最旧的消息：%v", oldClientConnState)
					}
					remoteAddr := ""
					if clientInfo.RemoteAddr != nil {
						remoteAddr = clientInfo.RemoteAddr.String()
					}
					s.ClientConnState <- fmsg.NewClientConnectState(clientInfo.ClientID, false, remoteAddr)
					s.log.Debugf("放入UnregisterClientID队列中的消息：%s", fmt.Sprintf("%+v", clientInfo))
				}
			case clientInfo, ok := <-s.hub.RegisterClientID:
				if ok {
					if len(s.ClientConnState) == cap(s.ClientConnState) {
						oldClientConnState := <-s.ClientConnState //去掉旧的，保留最新的,以防缓存满后，造成阻塞
						s.log.Debugf("ClientConnState消息队列已满，删除最旧的消息：%v", oldClientConnState)
					}
					remoteAddr := ""
					if clientInfo.RemoteAddr != nil {
						remoteAddr = clientInfo.RemoteAddr.String()
					}

					s.ClientConnState <- fmsg.NewClientConnectState(clientInfo.ClientID, true, remoteAddr)
					s.log.Debugf("放入RegisterClientID队列中的消息：%s", fmt.Sprintf("%+v", clientInfo))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return doneChan
}

// 获取当前在线的客户端ID列表
func (s *Server) GetOnlineClientIDs() []string {
	defer func() {
		if e := recover(); e != nil {
			err := errors.New(fmt.Sprintln(e))
			s.log.Errorf("%v", err) //输出到致命（关键错误）日志跟踪器
		}
	}()

	return s.hub.GetOnlineClientIDs()
}

// 关闭指定ID的websocket客户端
func (s *Server) CloseClient(ctx context.Context, wsClientID string) (err error) {
	if s.hub == nil {
		return errors.New("h.hub为空！")
	}
	return s.hub.CloseClient(ctx, wsClientID)
}

type ioMessage struct {
	Client *connClient //客户端
	data   []byte      //数据
}
