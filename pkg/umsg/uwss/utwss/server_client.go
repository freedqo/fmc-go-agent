// server_client_hub
package utwss

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net"
	"time"
)

// 构建服务端接入的客户端
func newConnClient(id string, opt *Option, hub *hub, conn *websocket.Conn, log *zap.SugaredLogger) *connClient {
	srv := &connClient{
		Id:   id,
		opt:  opt,
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 1024), //允许1024个缓冲
		log:  log,
	}
	srv.conn.SetPingHandler(srv.PingHandler)
	srv.conn.SetPongHandler(srv.PongHandler)
	return srv
}

// 服务端接入的客户端
type connClient struct {
	Id   string          //客户端ID
	opt  *Option         //配置
	hub  *hub            //客户端管理中心
	conn *websocket.Conn // The websocket connection.
	send chan []byte     // Buffered channel of outbound messages.
	log  *zap.SugaredLogger
}

// 启动服务
func (c *connClient) startService(ctx context.Context) {
	c.readPump(ctx)
	c.writePump(ctx)
}

// 从连接读取消息到客户端管理中心
func (c *connClient) readPump(ctx context.Context) (done <-chan struct{}) {
	doneChan := make(chan struct{})
	go func() {
		defer func() {
			if e := recover(); e != nil {
				c.log.Errorf("%v", e) //输出到致命（关键错误）日志跟踪器
			}
		}()
		defer close(doneChan)
		defer func() {
			var remoteAddr net.Addr = nil
			if c.conn != nil {
				remoteAddr = c.conn.RemoteAddr()
			}
			c.hub.unregister <- &serverClientWithRemoteAddr{
				ClientID:        c.Id,
				ServerClientObj: c,
				RemoteAddr:      remoteAddr,
			}
			c.conn.Close()
			c.log.Debugf("%v ←→ %v closed", c.conn.LocalAddr(), c.conn.RemoteAddr())
		}()
		c.conn.SetReadLimit(c.opt.MaxMessageSize)
		c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.opt.PongWaitSecond) * time.Second))
		subDone := make(chan struct{})
		go func() {
			defer func() {
				if e := recover(); e != nil {
					c.log.Errorf("%v", e) //输出到致命（关键错误）日志跟踪器
				}
			}()
			defer close(subDone)
			for {
				c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.opt.PongWaitSecond) * time.Second))
				msgType, message, err := c.conn.ReadMessage()
				if err != nil {
					// 1000是正常关闭，1001是客户端离开，均为预期事件
					if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
						c.log.Debugf("%v ←→ %v 正常关闭：%v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
					} else if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
						// 非预期关闭（如异常断开）才记为ERROR
						c.log.Errorf("%v ←→ %v 异常关闭：%v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
					} else if errors.Is(err, net.ErrClosed) {
						_ = c.conn.Close()
						break
					} else {
						// 其他错误（如网络问题）记为DEBUG或ERROR（根据业务需求）
						c.log.Errorf("%v ←→ %v 关闭：%v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
					}
					break
				}
				c.log.Debugf("%v ←→ %v 收到消息 类型：%v 内容：%s", c.conn.LocalAddr(), c.conn.RemoteAddr(), msgType, string(message))
				var msg ioMessage
				msg.Client = c
				msg.data = make([]byte, 0, len(message))
				msg.data = append(msg.data, message...)
				c.hub.iMsg <- msg
			}
		}()
		select {
		case <-subDone:
		case <-ctx.Done():
		}
	}()
	return doneChan
}

// 将客户端管理中心传入的消息写入到连接
func (c *connClient) writePump(ctx context.Context) (done <-chan struct{}) {
	doneChan := make(chan struct{})
	go func() {
		defer func() {
			if e := recover(); e != nil {
				c.log.Errorf("%v", e) //输出到致命（关键错误）日志跟踪器
			}
		}()
		ticker := time.NewTicker(c.opt.PingPeriod())
		if !c.opt.IsPing {
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
				case message, ok := <-c.send:
					c.log.Debugf("%v ←→ %v 开始发送消息长度：%d", c.conn.LocalAddr(), c.conn.RemoteAddr(), len(message))
					c.conn.SetWriteDeadline(time.Now().Add(time.Duration(c.opt.WriteWaitSecond) * time.Second))
					if !ok {
						c.conn.WriteMessage(websocket.CloseMessage, []byte{})
						c.log.Debugf("%v ←→ %v websocket.CloseMessage", c.conn.LocalAddr(), c.conn.RemoteAddr())
						return
					}

					w, err := c.conn.NextWriter(websocket.TextMessage)
					if err != nil {
						c.log.Errorf("%v ←→ %v %v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
						return
					}
					_, err = w.Write(message)
					if err == nil {
						c.log.Debugf("%v ←→ %v 发送成功！", c.conn.LocalAddr(), c.conn.RemoteAddr())
					}

					if err := w.Close(); err != nil {
						c.log.Errorf("%v ←→ %v 关闭MessageWriter错误：%v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
						return
					}
				case <-ticker.C:
					c.conn.SetWriteDeadline(time.Now().Add(time.Duration(c.opt.WriteWaitSecond) * time.Second))
					if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						c.log.Errorf("%v ←→ %v websocket.PingMessage错误：%v", c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
						return
					}
					c.log.Debugf("%v ←→ %v websocket.PingMessage", c.conn.LocalAddr(), c.conn.RemoteAddr())
				}
			}
		}()
		select {
		case <-subDone:
		case <-ctx.Done():
		}
	}()
	return doneChan
}

func (c *connClient) PingHandler(appData string) error {
	defer func() {
		if e := recover(); e != nil {
			c.log.Errorf("%v", e) //输出到致命（关键错误）日志跟踪器
		}
	}()
	c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.opt.PongWaitSecond) * time.Second))
	err := c.conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Duration(c.opt.WriteWaitSecond)*time.Second))
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

func (c *connClient) PongHandler(appData string) error {
	c.log.Debugf("%s", appData)
	c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.opt.PongWaitSecond) * time.Second))
	return nil
}

// 主动断开连接
func (c *connClient) Close(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			c.log.Errorf("connClient%s%s%v", "Close()", "err=", e)
		}
	}()
	if c.conn == nil {
		return
	}
	var remoteAddr net.Addr = nil
	if c.conn != nil {
		remoteAddr = c.conn.RemoteAddr()
	}
	c.hub.unregister <- &serverClientWithRemoteAddr{
		ClientID:        c.Id,
		ServerClientObj: c,
		RemoteAddr:      remoteAddr,
	}
	c.conn.Close()
	c.log.Debugf("%v ←→ %v closed", c.conn.LocalAddr(), c.conn.RemoteAddr())
	c.conn = nil
}
