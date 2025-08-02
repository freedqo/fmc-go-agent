// server_client_hub
package utwss

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/msgm"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"

	"go.uber.org/zap"
	"net"
	"strings"
	"sync"
)

// 构建服务端接入的客户端的管理中心
func newHub(log *zap.SugaredLogger) *hub {
	return &hub{
		register:           make(chan *serverClientWithRemoteAddr),
		unregister:         make(chan *serverClientWithRemoteAddr),
		iMsg:               make(chan ioMessage, 1024),
		clients:            make(map[*connClient]bool),
		ClientidToClients:  make(map[string]*connClient),
		ClientToClientids:  make(map[*connClient]string),
		ReceiveMsg:         make(chan fmsg.InMessage, 1024),
		SendMsg:            make(chan fmsg.OutMessage, 1024),
		UnregisterClientID: make(chan *registerClientInfo, 1024),
		RegisterClientID:   make(chan *registerClientInfo, 1024),
		log:                log,
		mut:                sync.RWMutex{},
	}
}

// 服务端接入的客户端的管理中心
type hub struct {
	mut sync.RWMutex //锁
	log *zap.SugaredLogger

	clients           map[*connClient]bool   //接入的客户端集合
	ClientidToClients map[string]*connClient //客户端ID到客户端的映射
	ClientToClientids map[*connClient]string //客户端到客户端ID的映射

	iMsg       chan ioMessage       //管道：接收客户收到的消息
	ReceiveMsg chan fmsg.InMessage  //管道：对外输出接收到的消息
	SendMsg    chan fmsg.OutMessage //管道：接收外部请发送到客户端的消息

	register           chan *serverClientWithRemoteAddr //管道：客户端注册,底层注册
	unregister         chan *serverClientWithRemoteAddr //管道：客户端注销，底层注销
	UnregisterClientID chan *registerClientInfo         //已经注销的客户端ID，通知上层应用
	RegisterClientID   chan *registerClientInfo         //新注册的客户端ID，通知上层应用

}

// 异步运行服务
func (h *hub) run(ctx context.Context) (done <-chan struct{}) {
	funcName := "run()"
	doneChan := make(chan struct{})
	go func() {
		defer func() {
			if e := recover(); e != nil {
				h.log.Errorf("%v", e) //输出到致命（关键错误）日志跟踪器
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case client := <-h.register:
				h.mut.Lock()
				h.clients[client.ServerClientObj] = true
				h.RegisterClientID <- &registerClientInfo{
					ClientID:   client.ClientID,
					RemoteAddr: client.RemoteAddr,
				}
				//建立客户端ID和客户端的映射
				_, ok := h.ClientToClientids[client.ServerClientObj]
				if !ok {
					h.ClientToClientids[client.ServerClientObj] = client.ClientID
					h.ClientidToClients[client.ClientID] = client.ServerClientObj
				}
				h.mut.Unlock()
			case client := <-h.unregister:
				h.mut.Lock()
				if _, ok := h.clients[client.ServerClientObj]; ok {
					delete(h.clients, client.ServerClientObj)
					close(client.ServerClientObj.send)
					clientid, ok := h.ClientToClientids[client.ServerClientObj]
					if ok {
						h.UnregisterClientID <- &registerClientInfo{
							ClientID:   clientid,
							RemoteAddr: client.RemoteAddr,
						}
						delete(h.ClientidToClients, clientid)
						delete(h.ClientToClientids, client.ServerClientObj)
					}
				}
				h.mut.Unlock()
			case omsg := <-h.SendMsg: //外部请求发送消息到客户端
				h.mut.Lock()
				if omsg.ClientID == "" {
					//广播发送
					for client := range h.clients {
						select {
						case client.send <- omsg.MessageData:
						default:
							close(client.send)
							delete(h.clients, client)
							clientid, ok := h.ClientToClientids[client]
							if ok {
								delete(h.ClientidToClients, clientid)
								delete(h.ClientToClientids, client)
							}
						}
					}

				} else if client, ok := h.ClientidToClients[omsg.ClientID]; ok {
					//发送到指定客户端
					select {
					case client.send <- omsg.MessageData:
					default:
						close(client.send)
						delete(h.clients, client)
						clientid, ok := h.ClientToClientids[client]
						if ok {
							delete(h.ClientidToClients, clientid)
							delete(h.ClientToClientids, client)
						}
					}
				}
				h.mut.Unlock()
			case imsg := <-h.iMsg: //从客户端接收到消息
				//解码出消息的基本信息
				var msgBase msgm.TAiAgentMessage
				err := json.Unmarshal(imsg.data, &msgBase)
				if err != nil {
					h.log.Errorw("解码错误",
						"func", funcName,
						"err", err,
						"消息内容", string(imsg.data))
					continue
				}
				// 强制刷新客户端对应的ClientID
				msgBase.SessionId = imsg.Client.Id

				//对外输出接收的消息
				inMsg, err := fmsg.NewInMessage(msgBase.MessageBase, msgBase.Data)
				if err == nil {
					if len(h.ReceiveMsg) == cap(h.ReceiveMsg) {
						<-h.ReceiveMsg //去掉旧的，保留最新的,以防缓存满后，造成阻塞
					}
					h.ReceiveMsg <- inMsg
				}
			}
		}
	}()
	return doneChan
}

// 获取当前在线的客户端ID列表
func (h *hub) GetOnlineClientIDs() []string {
	defer func() {
		if e := recover(); e != nil {
			err := errors.New(strings.TrimSpace(fmt.Sprintln(e)))
			h.log.Errorf("%v", err) //输出到致命（关键错误）日志跟踪器
		}
	}()

	h.mut.RLock()
	defer h.mut.RUnlock()
	ids := make([]string, 0)
	for id, _ := range h.ClientidToClients {
		ids = append(ids, id)
	}
	return ids
}

// 关闭指定ID的websocket客户端
func (h *hub) CloseClient(ctx context.Context, wsClientID string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			h.log.Errorw("hub CloseClient 错误", "err", e)
		}
	}()
	h.mut.Lock()
	defer h.mut.Unlock()
	client, isOk := h.ClientidToClients[wsClientID]
	if !isOk {
		return errors.New(fmt.Sprintf("websocket客户端[wsClientID=%v]不存在", wsClientID))
	}
	client.Close(ctx)
	return nil
}

// 指定ID的websocket客户端的地址
func (h *hub) GetClientRemoteAddr(wsClientID string) (remoteAddr net.Addr, err error) {
	defer func() {
		if e := recover(); e != nil {
			h.log.Errorw("hub GetClientRemoteAddr 错误", "err", e)
		}
	}()
	h.mut.Lock()
	defer h.mut.Unlock()
	client, isOk := h.ClientidToClients[wsClientID]
	if !isOk {
		return nil, errors.New(fmt.Sprintf("websocket客户端[wsClientID=%v]不存在", wsClientID))
	}
	if client.conn == nil {
		return nil, errors.New(fmt.Sprintf("websocket客户端[wsClientID=%v]，连接不存在", wsClientID))
	}
	remoteAddr = client.conn.RemoteAddr()
	return remoteAddr, nil
}

/*
	指定ID的websocket客户端的地址

remoteAddr - 字符串表示的客户地址。
*/
func (h *hub) GetClientRemoteAddrString(wsClientID string) (remoteAddr string) {
	defer func() {
		if e := recover(); e != nil {
			h.log.Errorw("hub GetClientRemoteAddrString 错误", "err", e)
		}
	}()
	remoteAddrNet, err := h.GetClientRemoteAddr(wsClientID)
	if err != nil {
		return ""
	}
	return remoteAddrNet.String()
}

type serverClientWithRemoteAddr struct {
	ClientID        string
	ServerClientObj *connClient
	RemoteAddr      net.Addr
}

type registerClientInfo struct {
	ClientID   string
	RemoteAddr net.Addr
}
