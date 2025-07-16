package uaimcpclient

import (
	"context"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaimcp"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaimcp/uaimcpclient/mcp2einotool"

	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

func New(ctx context.Context, name string, opt *uaimcp.Option) If {
	// 判断配置项是否为空
	if opt == nil {
		panic("MCP客户端配置项不能为空")
	}
	// 创建UAiMcpClient实例
	uMcpClient := &UAiMcpClient{
		name: name,
		opt:  opt,
	}
	var err error
	var mcpClient *client.Client
	// 根据不同的选项创建不同的客户端
	switch opt.Type {
	case uaimcp.MCP_Type_http_sse:
		{
			// 如果配置项中有OAuthConfig，则创建SSE传输客户端
			if opt.OAuthConfig != nil {
				// 创建 SSE 传输客户端
				mcpClient, err = client.NewOAuthSSEClient(opt.BaseURL, *opt.OAuthConfig)
				if err != nil {
					panic(err)
				}
			} else {
				// 创建 SSE 传输客户端
				mcpClient, err = client.NewSSEMCPClient(opt.BaseURL)
				if err != nil {
					panic(err)
				}
			}

		}
	case uaimcp.MCP_Type_streamable_http:
		{
			if opt.OAuthConfig != nil {
				// 创建 HTTP 状态保持传输客户端
				mcpClient, err = client.NewOAuthStreamableHttpClient(opt.BaseURL, *opt.OAuthConfig)
				if err != nil {
					panic(err)
				}
			} else {
				// 创建 HTTP 状态保持传输客户端
				mcpClient, err = client.NewStreamableHttpClient(opt.BaseURL)
				if err != nil {
					panic(err)
				}
			}
		}
	case uaimcp.MCP_Type_stdio:
		{
			// 创建 HTTP 状态保持传输客户端
			mcpClient, err = client.NewStdioMCPClient(opt.Command, opt.Env, opt.Args...)
			if err != nil {
				panic(err)
			}
		}
	default:
		{
			panic("不支持的MCP客户端类型")
		}
	}
	if mcpClient == nil {
		panic("MCP客户端创建失败")
	}
	uMcpClient.client = mcpClient

	err = mcpClient.Start(ctx)
	if err != nil {
		panic(err)
	}
	go func() {
		<-ctx.Done()
		err := mcpClient.Close()
		if err != nil {
			return
		}
	}()
	err = uMcpClient.Initialize(ctx)
	if err != nil {
		panic(err)
	}

	return uMcpClient
}

type UAiMcpClient struct {
	ctx        context.Context
	lCtx       context.Context
	lCancel    context.CancelFunc
	name       string
	version    string
	opt        *uaimcp.Option
	client     *client.Client
	ServerInfo *mcpgo.InitializeResult
}

func (u *UAiMcpClient) Start(ctx context.Context) (done <-chan struct{}, err error) {
	if u.client == nil {
		panic("MCP客户端未初始化")
	}
	u.ctx = ctx
	u.lCtx, u.lCancel = context.WithCancel(ctx)
	dc := make(chan struct{})
	err = u.client.Start(u.lCtx)
	if err != nil {
		return nil, err
	}
	return dc, err
}

func (u *UAiMcpClient) Stop() error {
	u.lCancel()
	err := u.client.Close()
	if err != nil {
		return err
	}
	return nil
}

func (u *UAiMcpClient) RestStart() (done <-chan struct{}, err error) {
	err = u.Stop()
	if err != nil {
		return nil, err
	}
	return u.Start(u.ctx)
}

func (u *UAiMcpClient) Close() error {
	if u.client == nil {
		panic("MCP客户端未初始化")
	}
	return u.client.Close()
}

func (u *UAiMcpClient) Initialize(ctx context.Context) error {
	initRequest := mcpgo.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcpgo.Implementation{
		Name:    u.name,
		Version: "1.0.0",
	}
	res, err := u.client.Initialize(ctx, initRequest)
	if err != nil {
		return err
	}
	u.ServerInfo = res
	u.name = res.ServerInfo.Name
	u.version = res.ServerInfo.Version
	return nil
}

func (u *UAiMcpClient) DToEinoTools(ctx context.Context) []tool.BaseTool {
	baseTools, err := mcp2einotool.GetTools(ctx, &mcp2einotool.Config{Cli: u.client})
	if err != nil {
		return nil
	}
	return baseTools
}

var _ If = &UAiMcpClient{}
