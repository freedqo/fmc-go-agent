package config

import (
	"github.com/freedqo/fmc-go-agents/pkg/fai/faicharmodel"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb"
	"github.com/freedqo/fmc-go-agents/pkg/fgrom"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg/umqt"
	"github.com/freedqo/fmc-go-agents/pkg/fzlog"
	"github.com/freedqo/fmc-go-agents/pkg/httpclient"
)

const (
	AppName = "fmc-go-agents-server"
)

func NewDefault() *Config {
	return &Config{
		Base: &BaseOption{
			Node:         1,
			HttpPort:     7589,
			TcpPort:      7588,
			LastExitTime: "",
		},
		Log: fzlog.NewDefaultOption(),
		Db:  fgrom.NewDefaultOption(),
		Ext: newExtOption(),
		UCM: &faicharmodel.Option{
			APIKey:       "",
			BaseURL:      "http://192.168.53.217:11434",
			Organization: "ut-pc2-gd",
			Provider:     "ollama",
			Model:        "deepseek-r1:7b",
			Timeout:      120,
		},
		MCP:     nil,
		Msg:     newMsgOption(),
		UVector: fvectordb.NewDefaultOption(),
	}
}

type Config struct {
	Base    *BaseOption          `comment:"基础配置"`
	Log     *fzlog.Option        `comment:"日志配置"`
	Db      *fgrom.Option        `comment:"数据库配置"`
	Ext     *ExtOption           `comment:"外部服务配置"`
	UCM     *faicharmodel.Option `comment:"UChatModel配置,用于连接大模型"`
	UVector *fvectordb.Option    `comment:"UVector,向量数据库检索"`
	MCP     *MCP                 `comment:"MCP配置"`
	Msg     *MsgOption           `comment:"消息配置"`
}

type BaseOption struct {
	Node         int    `comment:"节点编号"`
	HttpPort     int    `comment:"Http服务端口"`
	TcpPort      int    `comment:"Tcp服务端口"`
	LastExitTime string `comment:"上次退出时间"`
}

type ExtOption struct {
	UserCenter     *httpclient.Option `comment:"用户中心配置"`
	Knowledge      *httpclient.Option `comment:"知识库服务配置"`
	UserCenterRoot *UserCenterRoot    `comment:"用户中心管理员配置"`
}
type MCP struct {
	Server  *faimcp.Option            `comment:"MCP服务端配置"`
	Clients map[string]*faimcp.Option `comment:"MCP客户端配置"`
}
type UserCenterRoot struct {
	Username string `comment:"用户中心管理员用户名"`
	Password string `comment:"用户中心管理员密码,sm3加密后密码字符串"`
}

func newExtOption() *ExtOption {
	return &ExtOption{
		UserCenter: httpclient.NewDefaultOption(),
		UserCenterRoot: &UserCenterRoot{
			Username: "admin",
			Password: "sm3加密后字符串",
		},
	}
}

type MsgOption struct {
	//MainWss *uwss.Option `comment:"主消息服务配置"`
	Mqtt *umqt.Option `comment:"MQTT消息服务配置"`
}

func newMsgOption() *MsgOption {
	opt := &MsgOption{
		Mqtt: umqt.NewDefaultOption(),
	}
	return opt
}
