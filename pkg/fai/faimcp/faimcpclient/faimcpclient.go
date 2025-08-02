package faimcpclient

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp/faimcpclient/mcp2eino"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

func New(ctx context.Context, name string, opt *faimcp.Option, log *zap.SugaredLogger) (If, error) {
	// 判断配置项是否为空
	if opt == nil {
		return nil, errors.New("MCP客户端配置项不能为空")
	}
	// 创建UAiMcpClient实例
	uMcpClient := &FAiMcpClient{
		name:    name,
		opt:     opt,
		midFunc: make(map[*mcp2eino.If]mcp2eino.If),
		log:     log,
	}
	var err error
	var mcpClient *client.Client
	// 根据不同的选项创建不同的客户端
	switch opt.Type {
	case faimcp.MCP_Type_http_sse:
		{
			// 如果配置项中有OAuthConfig，则创建SSE传输客户端
			if opt.OAuthConfig != nil {
				// 创建 SSE 传输客户端
				mcpClient, err = client.NewOAuthSSEClient(opt.BaseURL, *opt.OAuthConfig)
				if err != nil {
					return nil, err
				}
			} else {
				// 创建 SSE 传输客户端
				mcpClient, err = client.NewSSEMCPClient(opt.BaseURL)
				if err != nil {
					return nil, err
				}
			}

		}
	case faimcp.MCP_Type_streamable_http:
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
					return nil, err
				}
			}
		}
	case faimcp.MCP_Type_stdio:
		{
			// 创建 HTTP 状态保持传输客户端
			mcpClient, err = client.NewStdioMCPClient(opt.Command, opt.Env, opt.Args...)
			if err != nil {
				return nil, err
			}
		}
	default:
		{
			return nil, errors.New("未知的MCP客户端类型")
		}
	}
	if mcpClient == nil {
		return nil, err
	}
	uMcpClient.client = mcpClient

	err = mcpClient.Start(ctx)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return uMcpClient, nil
}

type FAiMcpClient struct {
	ctx        context.Context
	lCtx       context.Context
	lCancel    context.CancelFunc
	name       string
	version    string
	opt        *faimcp.Option
	client     *client.Client
	serverInfo *mcp.InitializeResult
	midFunc    map[*mcp2eino.If]mcp2eino.If
	log        *zap.SugaredLogger
}

func (u *FAiMcpClient) SubToolMidFunc(fun *mcp2eino.If) {
	u.midFunc[fun] = *fun
}

func (u *FAiMcpClient) UnSubToolMidFunc(fun *mcp2eino.If) {
	delete(u.midFunc, fun)
}

func (u *FAiMcpClient) Start(ctx context.Context) (done <-chan struct{}, err error) {
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

func (u *FAiMcpClient) Stop() error {
	u.lCancel()
	err := u.client.Close()
	if err != nil {
		return err
	}
	return nil
}

func (u *FAiMcpClient) RestStart() (done <-chan struct{}, err error) {
	err = u.Stop()
	if err != nil {
		return nil, err
	}
	return u.Start(u.ctx)
}

func (u *FAiMcpClient) Close() error {
	if u.client == nil {
		panic("MCP客户端未初始化")
	}
	return u.client.Close()
}

func (u *FAiMcpClient) Initialize(ctx context.Context) error {
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{}

	res, err := u.client.Initialize(ctx, initRequest)
	if err != nil {
		return err
	}
	u.serverInfo = res
	u.name = res.ServerInfo.Name
	u.version = res.ServerInfo.Version
	tools, err := u.client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return err
	}
	for _, t := range tools.Tools {
		fmt.Printf("加载[%s:%s]MCP服务工具:%s，Annotations.Title:%s\r\n", u.name, u.version, t.Name, t.Annotations.Title)
	}
	return nil
}
func (u *FAiMcpClient) ServerInfo() *mcp.InitializeResult {
	return u.serverInfo
}

func (u *FAiMcpClient) DToEinoTools(ctx context.Context) []tool.BaseTool {
	baseTools, err := mcp2eino.GetTools(ctx, &mcp2eino.Config{
		Cli:     u.client,
		MidFunc: u.midFunc,
	},
	)
	if err != nil {
		return nil
	}
	return baseTools
}

func (u *FAiMcpClient) CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return u.client.CallTool(ctx, request)
}

var _ If = &FAiMcpClient{}
