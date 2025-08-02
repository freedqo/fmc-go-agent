/* server_cfg
 * 优化：自定义序列化 hws 2022.08.12
 */

package utwss

import (
	"time"
)

// utwebsecket服务端配置
type Option struct {
	Enable          bool   `comment:"是否启用"`               //是否启用
	WriteWaitSecond int    `comment:"写数据超时[秒]"`           //写数据超时[秒]
	PongWaitSecond  int    `comment:"读数据超时[秒]"`           //读数据超时[秒]
	MaxMessageSize  int64  `comment:"读数据最大缓冲区;0-为没有限制"`   //读数据最大缓冲区;0-为没有限制
	ServicePort     int    `comment:"服务侦听端口"`             //服务侦听端口
	ServicePattern  string `comment:"服务路由目录"`             //服务路由目录
	IsPing          bool   `comment:"是否从本服务端给客户端定时发ping"` //是否从本服务端给客户端定时发ping
}

// 构建utwebsecket服务端配置
func NewDefaultOption(port int) Option {
	cfg := Option{
		WriteWaitSecond: 30,
		PongWaitSecond:  120,
		MaxMessageSize:  0, //不限制信息大小
		ServicePort:     port,
		ServicePattern:  "/ws",
		IsPing:          true,
	}
	return cfg
}

// 获取ping间隔秒数
func (s Option) PingPeriod() time.Duration {
	return (time.Duration(s.PongWaitSecond) * time.Second * 9) / 16
}
