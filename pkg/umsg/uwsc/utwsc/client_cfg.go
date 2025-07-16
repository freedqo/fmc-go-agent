/* client_cfg - utwebsocket客户端配置
 * 优化：自定义序列化 hws 2022.08.12
 */

package utwsc

import (
	"fmt"
	"time"
)

// utwebsocket客户端配置
type Option struct {
	RemoteIP1              string `xml:"RemoteIP1"`              //远方IP或主机(主)
	RemoteIP2              string `xml:"RemoteIP2"`              //远方IP或主机(备)(如果不为空，则拨号连接时，如不拨号失败，将在RemoteIP1和RemoteIP2之间轮询拨号)
	RemotePort             int    `xml:"RemotePort"`             //远方端口
	RemotePath             string `xml:"RemotePath"`             //远方服务路径
	WriteWaitSecond        int    `xml:"WriteWaitSecond"`        //写数据超时秒数
	HandshakeTimeoutSecond int    `xml:"HandshakeTimeoutSecond"` //握手超时时间秒数
	RedialIntervalSecond   int    `xml:"RedialIntervalSecond"`   //重拨间隔秒数。（即拨号失败后，间隔多长时间后再次重新拨号）
	IsPing                 bool   `xml:"IsPing"`                 //是否从本客户端给服务端定时发ping
	PongWaitSecond         int    `xml:"PongWaitSecond"`         //读数据超时
}

// 构建utwebsocket客户端配置(使用默认参数)
func NewOption() Option {
	cfg := NewClientCfg2("127.0.0.1", "", 5001, "/ws")
	return cfg
}

// 构建utwebsocket客户端配置
func NewClientCfg2(remoteIP1 string, remoteIP2 string, remotePort int, remotePath string) Option {
	cfg := Option{RemoteIP1: remoteIP1,
		RemoteIP2:              remoteIP2,
		RemotePort:             remotePort,
		RemotePath:             remotePath,
		WriteWaitSecond:        10,
		HandshakeTimeoutSecond: 45,
		RedialIntervalSecond:   5,
		PongWaitSecond:         60,
		IsPing:                 true,
	}
	return cfg
}

func (c *Option) String() string {
	return fmt.Sprintf("{RemoteIP1=%s,RemoteIP2=%s,RemotePort=%d,RemotePath=%s,WriteWaitSecond=%d,HandshakeTimeoutSecond=%d,RedialIntervalSecond=%d}",
		c.RemoteIP1, c.RemoteIP2, c.RemotePort, c.RemotePath, c.WriteWaitSecond, c.HandshakeTimeoutSecond, c.RedialIntervalSecond)
}

// 获取ping间隔秒数
func (c Option) PingPeriod() time.Duration {
	return (time.Duration(c.PongWaitSecond) * time.Second * 9) / 16
}
