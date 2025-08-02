package faimcp

import "github.com/mark3labs/mcp-go/client"

const (
	MCP_Type_http_sse        string = "sse"
	MCP_Type_streamable_http string = "streamable"
	MCP_Type_stdio           string = "stdio"
)

type Option struct {
	Type        string              `comment:"传输协议类型,sse、streamable、stdio" json:"type"`
	Command     string              `comment:"stdio传输协议的命令" json:"command"`
	Env         []string            `comment:"stdio传输协议的环境变量" json:"env"`
	Args        []string            `comment:"stdio传输协议的命令参数" json:"args"`
	BaseURL     string              `comment:"sse、streamable传输协议的远程链接,客户端配置到具体端点:http://localhost:7856/mcp,服务端配置到暴露端口：0.0.0.0:7856" json:"baseURL"`
	Header      map[string]string   `comment:"sse、streamable传输协议的请求头" json:"header"`
	OAuthConfig *client.OAuthConfig `comment:"sse、streamable传输协议的OAuth配置" json:"oauthConfig"`
}

func NewDefaultOption() *Option {
	return &Option{
		Type:    MCP_Type_streamable_http,
		Command: "ls",
		Env:     []string{"test"},
		Args:    []string{"test"},
		BaseURL: "http://localhost:7856/mcp",
		Header: map[string]string{
			"Authorization": "Bearer your-auth-token", // 正确格式：映射类型
			"Content-Type":  "application/json",
		},
		OAuthConfig: nil,
	}
}
