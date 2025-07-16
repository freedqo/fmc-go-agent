package utwsc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	receiveMsgSize = 1024 //接收消息的缓冲数量
	sendMsgSize    = 1024 //发送消息的缓冲数量
)

// 构建utwebsocket客户端
func NewClient() *Client {
	cfg := NewOption()
	client := &Client{Cfg: &cfg,
		dialer:          &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: time.Duration(cfg.HandshakeTimeoutSecond) * time.Second},
		conn:            nil,
		ReceiveMsg:      make(chan umsg.InMessage, receiveMsgSize),
		sendMsg:         make(chan []byte, sendMsgSize),
		ClientConnState: make(chan umsg.ClientConnectState, 1024),
	}
	return client
}

// utwebsocket客户端
type Client struct {
	Cfg                    *Option
	dialer                 *websocket.Dialer            //
	conn                   *websocket.Conn              //当前连接
	mutConn                sync.RWMutex                 //conn的锁
	ReceiveMsg             chan umsg.InMessage          //管道：对外输出接收的消息。如果外部不及时取出接收的这些消息，达到最大缓冲后将丢掉旧的消息。
	sendMsg                chan []byte                  //管道：接收外部请发送到客户端的消息
	cancel                 context.CancelFunc           //取消本服务的方法
	parentCtx              context.Context              //调用本服务的context
	curDailFailIP          string                       //当前拨号失败IP
	mutCurDailFailIP       sync.RWMutex                 //curDailFailIP的锁
	cnnState               bool                         //当前连接状态。true-连接;false-断开
	mutCnnState            sync.RWMutex                 //cnnState的锁
	clientGuid             string                       //客户端ID
	mutclientGuid          sync.RWMutex                 //clientGuid的锁
	ClientConnState        chan umsg.ClientConnectState //管道：向外推送客户端的连接状态事件
	once                   sync.Once
	pos                    string            //结构体位置描述
	isTransparentTransport bool              //是否透明传输。即websocket客户端和服务端数据传输，不遵守Message格式。
	log                    zap.SugaredLogger //专用日志
}

// 启动服务
// clientID - 客户端Guid。
func (c *Client) StartService(ctx context.Context, cfg Option, clientID string) (done <-chan struct{}, err error) {
	return c.startService(ctx, cfg, clientID, false)
}

/*
启动服务
clientID - 客户端Guid。
isTransparentTransport - 是否透明传输。即websocket客户端和服务端数据传输，不遵守Message格式。
*/
func (c *Client) StartService2(ctx context.Context, cfg Option, clientID string, isTransparentTransport bool) (done <-chan struct{}, err error) {
	return c.startService(ctx, cfg, clientID, isTransparentTransport)
}

// 启动服务
// clientID - 客户端Guid。
func (c *Client) startService(ctx context.Context, cfg Option, clientID string, isTransparentTransport bool) (done <-chan struct{}, err error) {

	if c.cancel != nil {
		c.cancel()
		time.Sleep(1 * time.Second)
	}
	c.isTransparentTransport = isTransparentTransport
	c.parentCtx = ctx
	c.Cfg = &cfg
	c.setClientGuid(clientID)
	myctx, cancel := context.WithCancel(c.parentCtx)
	c.cancel = cancel
	c.log.Debugf("启动连接远程服务：%v", cfg)
	return c.beginLinkService(myctx), nil
}

// 单独停止本服务
// 也可以通过Context，在统一结束所有服务时，结束本服务。
func (c *Client) StopService() (err error) {
	//funcName := "StopService()"
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintln(e))
			c.log.Errorf("%v", err) //输出到致命（关键错误）日志跟踪器
		}
	}()
	c.log.Errorf("请求结束服务: %v", c.Cfg)
	c.cancel()
	return nil
}

// 重启动服务
// 调用RestartService()方法之前，请先结束服务（例如，可以调用StopService()方法等）
// 基于上次调用StartService()方法的Context和Cfg启动服务。
// 如果之前没有调用过StartService()方法，则将会返回错误
// 如果Cfg有变化，则先修改Cfg,再调用本方法
// 也可以直接调用StartService()方法重新启动服务。但要注意的是：要正确使用管理本服务Context，以确保外部能够通过统一方式正确结束所有服务（包括本服务）。
func (c *Client) RestartService() (done <-chan struct{}, err error) {
	//funcName := "RestartService()"
	c.log.Debugf("连接配置：%v 请求重启服务", c.Cfg)
	if c.parentCtx == nil {
		err = errors.New("之前没调用过StartService()方法！")
		c.log.Debugf("连接配置：%v 请求重启服务错误：%v", c.Cfg, err)
		return done, err
	}
	return c.startService(c.parentCtx, *c.Cfg, c.clientGuid, c.isTransparentTransport)
}

// 发送消息。异步发送，无法确保消息发送到服务端。
// msg.ClientID - 为本地客户端GUID。一般情况下，要和StartService（）方法中的clientID参数保持一致，运行过程不会改变
func (c *Client) SendMessage(msg umsg.Message) (err error) {
	//funcName := "SendMessage()"
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintln(e))
			c.log.Errorf("%v", err) //输出到致命（关键错误）日志跟踪器
		}
	}()
	c.log.Debugf("请求发送消息：%v", msg)
	if !c.IsConnected() {
		err = errors.New("客户端未连接到服务器！")
		c.log.Debugf("SendMessage错误：%v", err)
		return err
	}
	if len(c.sendMsg) == cap(c.sendMsg) {
		err = errors.New("发送消息队列已满")
		c.log.Debugf("SendMessage错误：%v", err)
		return err
	}
	//构建输出消息
	if c.isTransparentTransport {
		// 先将该消息进行业务处理。（可选）
		if msg.OperateDataType != umsg.TransparentTransport_package_message_OperateDataType || msg.OperateData == nil {
			return errors.New("当前为透明传输方式，但是消息为非透明方式，或者透明数据为空！")
		}
		msgJsonData, err := json.Marshal(msg)
		if err != nil {
			return errors.New(fmt.Sprintf("json.Marshal(msg) 序列化失败:%v", err))
		}
		var message umsg.Message
		realOperateData := make([]byte, 0)
		message.OperateData = &realOperateData
		err = json.Unmarshal(msgJsonData, &message)
		if err != nil {
			return errors.New(fmt.Sprintf("json.Unmarshal(msg) 反序列化失败:%v", err))
		}
		outMsg := umsg.OutMessage{ClientID: "",
			MessageData: realOperateData,
		}
		c.sendMsg <- outMsg.MessageData
	} else {
		//构建输出消息
		outMsg, err := umsg.NewOutMessage(msg)
		if err != nil {
			return err
		}
		c.setClientGuid(outMsg.ClientID)
		c.sendMsg <- outMsg.MessageData
	}

	return nil
}

// 设置客户端ID
func (c *Client) setClientGuid(clientID string) {
	c.mutclientGuid.Lock()
	defer c.mutclientGuid.Unlock()
	c.clientGuid = clientID
}

// 获取客户端ID
func (c *Client) getClientGuid() string {
	c.mutclientGuid.RLock()
	defer c.mutclientGuid.RUnlock()
	return c.clientGuid
}

// 当前是否连接
func (c *Client) IsConnected() bool {
	c.mutCnnState.RLock()
	defer c.mutCnnState.RUnlock()
	return c.cnnState
}

// 设置连接状态
func (c *Client) setCurConnected(isConnected bool) {
	//funcName := "setCurConnected()"
	c.mutCnnState.Lock()
	defer c.mutCnnState.Unlock()
	if c.cnnState != isConnected {
		c.cnnState = isConnected
		if len(c.ClientConnState) == cap(c.ClientConnState) {
			oldClientConnState := <-c.ClientConnState //去掉旧的，保留最新的,以防缓存满后，造成阻塞
			c.log.Debugf("ClientConnState消息队列已满，删除最旧的消息：%v", oldClientConnState)
		}
		c.ClientConnState <- umsg.NewClientConnectState(c.clientGuid, isConnected, c.GetRemoteAddrString())
	}
}

// 关闭连接
func (c *Client) closeConn() {
	//funcName := "closeConn()"
	c.mutConn.Lock()
	defer c.mutConn.Unlock()
	if c.conn != nil {
		c.log.Debugf("%v ←→ %v 断开连接", c.conn.LocalAddr(), c.conn.RemoteAddr())
		c.conn.Close()
		c.conn = nil
	}
	c.setCurConnected(false)
}

// 关闭连接
func (c *Client) setConnected(cnn *websocket.Conn) {
	//funcName := "setConnected()"
	c.mutConn.Lock()
	defer c.mutConn.Unlock()
	if c.conn != nil {
		c.log.Debugf("%v ←→ %v 断开连接", c.conn.LocalAddr(), c.conn.RemoteAddr())
		c.conn.Close()
		c.conn = nil
	}
	c.conn = cnn
	c.conn.SetPingHandler(c.PingHandler)
	c.conn.SetPongHandler(c.PongHandler)
	c.setCurConnected(true)
}

// 获取下一个拨号IP(双网处理)
func (c *Client) getNextDialIP() string {
	c.mutCurDailFailIP.RLock()
	defer c.mutCurDailFailIP.RUnlock()
	if c.curDailFailIP == "" {
		return c.Cfg.RemoteIP1
	}
	if c.Cfg.RemoteIP1 == c.curDailFailIP && c.Cfg.RemoteIP2 != "" {
		return c.Cfg.RemoteIP2
	}
	return c.Cfg.RemoteIP1
}

// 设置当前拨号失败IP
func (c *Client) setCurDailFailIP(ip string) {
	c.mutCurDailFailIP.Lock()
	defer c.mutCurDailFailIP.Unlock()
	c.curDailFailIP = ip
}

// 开始连接服务
func (c *Client) beginLinkService(ctx context.Context) (done <-chan struct{}) {
	doneChan := make(chan struct{})
	go func() {
		defer func() {
			if e := recover(); e != nil {
				c.log.Errorf("%v", e) //输出到致命（关键错误）日志跟踪器
			}
		}()
		defer close(doneChan)
		defer c.closeConn()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			subDone, err := c.dialAndServiceOnce(ctx)
			if err != nil {
				c.closeConn()
				time.Sleep(time.Duration(rand.Int()%(c.Cfg.RedialIntervalSecond+1)) * time.Second)
				continue
			}
			select {
			case <-subDone:
			case <-ctx.Done():
				return
			}
		}
	}()
	return doneChan
}

// 建立连接，并提供服务
// 如果连接成功，则提供发送数据和接收数据服务，直到连接断开为止
func (c *Client) dialAndServiceOnce(ctx context.Context) (done <-chan struct{}, err error) {
	//funcName := "dialAndServiceOnce()"
	doneChan := make(chan struct{})
	defer func() {
		if e := recover(); e != nil {
			close(doneChan)
			err = errors.New(fmt.Sprintln(e))
			c.log.Errorf("%v", err) //输出到致命（关键错误）日志跟踪器
		}
	}()
	//开始拨号
	nextDialIP := c.getNextDialIP()
	c.dialer.HandshakeTimeout = time.Duration(c.Cfg.HandshakeTimeoutSecond) * time.Second
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", nextDialIP, c.Cfg.RemotePort), Path: c.Cfg.RemotePath}
	c.log.Debugf("开始连接：%s", u.String())
	cnn, _, err := c.dialer.Dial(u.String(), nil)
	if err != nil {
		c.log.Debugf("连接失败：%v", err)
		c.setCurDailFailIP(nextDialIP)
		close(doneChan)
		return doneChan, err
	}
	c.setConnected(cnn)
	c.log.Debugf("%v ←→ %v 连接成功！", cnn.LocalAddr(), cnn.RemoteAddr())

	//启动读写服务
	go func() {
		defer func() {
			if e := recover(); e != nil {
				c.log.Errorf("%v", err) //输出到致命（关键错误）日志跟踪器
			}
		}()
		defer close(doneChan)
		subDone := make(chan struct{})
		go func() {
			defer func() {
				if e := recover(); e != nil {
					c.log.Errorf("%v", err) //输出到致命（关键错误）日志跟踪器
				}
			}()
			defer close(subDone)
			var wg sync.WaitGroup
			subCtx, subCancel := context.WithCancel(ctx)
			readDone := c.beginReadAllMessage(subCtx)
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-readDone
				subCancel()
			}()
			writeDone := c.beginWriteAllMessage(subCtx)
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-writeDone
				subCancel()
			}()
			wg.Wait()
		}()
		//采用Message格式消息，要发送握手
		if !c.isTransparentTransport {
			msg := umsg.NewMessageForSend(c.getClientGuid(), "", "", nil)
			omsg, err2 := umsg.NewOutMessage(msg)
			if err2 == nil {
				c.sendMsg <- omsg.MessageData
			}
		}
		<-subDone
	}()

	return doneChan, nil
}

// 开始异步读取所有消息
func (c *Client) beginReadAllMessage(ctx context.Context) (done <-chan struct{}) {
	//funcName := "beginReadAllMessage()"
	doneChan := make(chan struct{})
	go func() {
		defer func() {
			if e := recover(); e != nil {
				c.log.Errorf("异常%v", e) //输出到致命（关键错误）日志跟踪器
			}
		}()
		defer close(doneChan)
		subDone := make(chan struct{})
		go func() {
			defer func() {
				if e := recover(); e != nil {
					c.log.Errorf("异常%v", e) //输出到致命（关键错误）日志跟踪器
				}
			}()
			defer close(subDone)
			for {
				c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.Cfg.PongWaitSecond) * time.Second))
				msgType, message, err := c.conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						c.log.Debugf("%v ←→ %v ReadMessage()错误：%v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
					}
					c.log.Debugf("%v ←→ %v ReadMessage() %v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
					return
				}
				c.log.Debugf("%v ←→ %v 收到消息 类型：%v 内容长度：%d", c.conn.LocalAddr(), c.conn.RemoteAddr(), msgType, len(message))
				if c.isTransparentTransport {
					//透明传输（非Message格式消息）,打包成通用的Message消息格式,然后再对外输出
					packageMsg := umsg.NewMessageForPackage(message)
					packageMsgData, err := json.Marshal(packageMsg)
					if err != nil {
						c.log.Debugf("%v ←→ %v 解码错误：%v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
						continue
					}
					var msgBase umsg.MessageBase = packageMsg.MessageBase //复制
					inMsg, err := umsg.NewInMessage(msgBase, packageMsgData)
					if err == nil {
						if len(c.ReceiveMsg) == cap(c.ReceiveMsg) {
							<-c.ReceiveMsg //去掉旧的，保留最新的,以防缓存满后，造成阻塞
						}
						c.ReceiveMsg <- inMsg
					}
				} else {
					//Message格式的消息,解码出消息的基本信息
					var msgBase umsg.MessageBase
					err = json.Unmarshal(message, &msgBase)
					if err != nil {
						c.log.Debugf("%v ←→ %v 解码错误：%v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
						continue
					}
					//对外输出接收的消息
					inMsg, err := umsg.NewInMessage(msgBase, message)
					if err == nil {
						if len(c.ReceiveMsg) == cap(c.ReceiveMsg) {
							<-c.ReceiveMsg //去掉旧的，保留最新的,以防缓存满后，造成阻塞
						}
						c.ReceiveMsg <- inMsg
					}
				}
				time.Sleep(10 * time.Millisecond)
			}
		}()
		select {
		case <-subDone:
		case <-ctx.Done():
		}
	}()
	return doneChan
}

// 开始异步发送所有消息
func (c *Client) beginWriteAllMessage(ctx context.Context) (done <-chan struct{}) {
	//funcName := "beginWriteAllMessage()"
	doneChan := make(chan struct{})
	go func() {
		defer func() {
			if e := recover(); e != nil {
				c.log.Errorf("异常%v", e) //输出到致命（关键错误）日志跟踪器
			}
		}()
		defer close(doneChan)
		defer c.closeConn()
		ticker := time.NewTicker(c.Cfg.PingPeriod())
		if !c.Cfg.IsPing {
			ticker.Stop()
		}
		defer func() {
			ticker.Stop()
			c.conn.Close()
			c.log.Debugf("本地地址：%v ←→ %v closed", c.conn.LocalAddr(), c.conn.RemoteAddr())
		}()
		subDone := make(chan struct{})
		go func() {
			defer func() {
				if e := recover(); e != nil {
					c.log.Errorf("%v", e) //输出到致命（关键错误）日志跟踪器
				}
			}()
			defer close(subDone)
			for {
				select {
				case omsg, ok := <-c.sendMsg: //外部请求发送消息到服务端
					if !ok {
						continue
					}
					if !c.IsConnected() {
						n := len(c.sendMsg)
						for i := 0; i < n; i++ {
							<-c.sendMsg
						}
						return
					}
					c.log.Debugf("%v ←→ %v 开始发送消息：%s", c.conn.LocalAddr(), c.conn.RemoteAddr(), string(omsg))
					c.conn.SetWriteDeadline(time.Now().Add(time.Duration(c.Cfg.WriteWaitSecond) * time.Second))
					w, err := c.conn.NextWriter(websocket.TextMessage)
					if err != nil {
						c.log.Debugf("%v ←→ %v NextWriter错误：%v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
						return
					}
					_, err = w.Write(omsg)
					if err == nil {
						c.log.Debugf("%v ←→ %v 发送成功！", c.conn.LocalAddr(), c.conn.RemoteAddr())
					}
					if err := w.Close(); err != nil {
						c.log.Debugf("%v ←→ %v 关闭MessageWriter错误：%v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
						return
					}
				case <-ticker.C:
					c.conn.SetWriteDeadline(time.Now().Add(time.Duration(c.Cfg.WriteWaitSecond) * time.Second))

					if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						c.log.Debugf("%v ←→ %v websocket.PingMessage错误：%v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
						return
					}

					c.log.Debugf("%v ←→ %v websocket.PingMessage", c.conn.LocalAddr(), c.conn.RemoteAddr())
				}
			}
		}()
		select {
		case <-subDone:
		case <-ctx.Done(): //读取线程可能已关闭(表示连接已断开),则结束本线程
		}
	}()
	return doneChan
}

// 获取结构体位置描述
func (c *Client) getPos() string {
	c.once.Do(func() {
		c.pos = strings.TrimSpace(fmt.Sprintln("urtyg-core", "msbase", fmt.Sprintf("%T", c)))
	})
	return c.pos
}

func (c *Client) PingHandler(appData string) error {
	//funcName := "PingHandler()"
	defer func() {
		if e := recover(); e != nil {
			c.log.Errorf("%v", e) //输出到致命（关键错误）日志跟踪器
		}
	}()
	c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.Cfg.PongWaitSecond) * time.Second))
	err := c.conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Duration(c.Cfg.WriteWaitSecond)*time.Second))
	if err == websocket.ErrCloseSent {
		c.log.Debugf("ErrCloseSent: %v", err)
		return nil
	} else if e, ok := err.(net.Error); ok && e.Temporary() {
		c.log.Debugf("net.Error: %v", err)
		return nil
	} else if err != nil {
		c.log.Debugf("其它错误: %v", err)
	}
	c.log.Debugf("receive ping from %v %s", c.conn.RemoteAddr(), appData)
	return err
}

func (c *Client) PongHandler(appData string) error {
	//funcName := "PongHandler()"
	c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.Cfg.PongWaitSecond) * time.Second))
	c.log.Debugf("%s", appData)
	return nil
}

// 获取服务端的地址
func (c *Client) GetRemoteAddr() (remoteAddr net.Addr, err error) {
	defer func() {
		if e := recover(); e != nil {
			c.log.Debugf("Client err=%v", e) //输出到致命（关键错误）日志跟踪器
		}
	}()
	if c.conn == nil {
		return nil, errors.New("未连接")
	}
	remoteAddr = c.conn.RemoteAddr()
	return remoteAddr, nil
}

/*
	获取服务端的地址

remoteAddr - 字符串表示的客户地址。
*/
func (c *Client) GetRemoteAddrString() (remoteAddr string) {
	defer func() {
		if e := recover(); e != nil {
			c.log.Debugf("Client err=%v", e) //输出到致命（关键错误）日志跟踪器
		}
	}()
	remoteAddrNet, err := c.GetRemoteAddr()
	if err != nil {
		return ""
	}
	return remoteAddrNet.String()
}
