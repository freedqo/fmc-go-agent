package config

import (
	"github.com/freedqo/fmc-go-agent/pkg/httpclient"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaicharmodel"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaimcp"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaivectordb"
	"github.com/freedqo/fmc-go-agent/pkg/ugrom"
	"github.com/freedqo/fmc-go-agent/pkg/umsg/umqt"
	"github.com/freedqo/fmc-go-agent/pkg/umsg/uwss"
	"github.com/freedqo/fmc-go-agent/pkg/uzlog"
)

const (
	AppName = "fmc-go-agent-server"
)

func NewDefault() *Config {
	return &Config{
		Base: &BaseOption{
			Node:         1,
			HttpPort:     7589,
			TcpPort:      7588,
			LastExitTime: "",
		},
		Log: uzlog.NewDefaultOption(),
		Db:  ugrom.NewDefaultOption(),
		Ext: newExtOption(),
		UCM: &uaicharmodel.Option{
			APIKey:       "",
			BaseURL:      "http://192.168.53.217:11434",
			Organization: "fmc",
			Provider:     "ollama",
			Model:        "deepseek-r1:7b",
			Timeout:      120,
		},
		UiRv:      uaivectordb.NewOption(),
		McpServer: uaimcp.NewDefaultOption(),
		Msg:       newMsgOption(),
	}
}

type Config struct {
	Base      *BaseOption          `comment:"基础配置"`
	Log       *uzlog.Option        `comment:"日志配置"`
	Db        *ugrom.Option        `comment:"数据库配置"`
	Ext       *ExtOption           `comment:"外部服务配置"`
	UCM       *uaicharmodel.Option `comment:"UChatModel配置,用于连接大模型"`
	UiRv      *uaivectordb.Option  `comment:"UAIIRVector配置,用于意图识别的向量检索"`
	McpServer *uaimcp.Option       `comment:"MCP服务配置(MCP服务)"`
	Msg       *MsgOption           `comment:"消息配置"`
}

type BaseOption struct {
	Node         int    `comment:"节点编号"`
	HttpPort     int    `comment:"Http服务端口"`
	TcpPort      int    `comment:"Tcp服务端口"`
	LastExitTime string `comment:"上次退出时间"`
}

type ExtOption struct {
	UserCenter     *httpclient.Option         `comment:"用户中心配置"`
	McpServer      *map[string]*uaimcp.Option `comment:"MCP客户配置(MCP客户)"`
	UserCenterRoot *UserCenterRoot            `comment:"用户中心管理员配置"`
}
type UserCenterRoot struct {
	Username string `comment:"用户中心管理员用户名"`
	Password string `comment:"用户中心管理员密码,sm3加密后密码字符串"`
}

func newExtOption() *ExtOption {
	return &ExtOption{
		UserCenter: httpclient.NewDefaultOption(),
		McpServer:  nil,
		UserCenterRoot: &UserCenterRoot{
			Username: "admin",
			Password: "sm3加密后字符串",
		},
	}
}

type MsgOption struct {
	MainWss *uwss.Option `comment:"主消息服务配置"`
	Mqtt    *umqt.Option `comment:"MQTT消息服务配置"`
}

func newMsgOption() *MsgOption {
	opt := &MsgOption{
		MainWss: uwss.NewDefaultOption(7896),
		Mqtt:    umqt.NewDefaultOption(),
	}
	return opt
}
