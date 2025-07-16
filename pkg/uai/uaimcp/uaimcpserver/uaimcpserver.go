package uaimcpserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaimcp"
	"github.com/freedqo/fmc-go-agent/pkg/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
	"net/http"
)

func New(ctx context.Context, opt *uaimcp.Option, log *zap.SugaredLogger, name, version string) If {
	if name == "" {
		name = "fcm-mcp-server:" + utils.GetStringID()
	}
	if version == "" {
		version = "1.0.0"
	}
	us := &UAiMcpServer{
		ctx:     ctx,
		name:    name,
		version: version,
		log:     log,
		opt:     opt,
	}
	us.server = server.NewMCPServer(name, version,
		us.WithRecovery(), // 添加一个中间件，用于在工具处理程序中恢复
		server.WithResourceCapabilities(true, true), //还不知道是干嘛的
		server.WithPromptCapabilities(true),         // 还不知道是干嘛的
		server.WithToolCapabilities(true),           // 还不知道是干嘛的
		server.WithHooks(us.newHooks()),             // 注册勾子方法，监控执行过程
		// 添加一个日志中间件，用于记录请求和响应
	)
	us.streamableSrv = server.NewStreamableHTTPServer(us.server)
	us.sseSrv = server.NewSSEServer(us.server)

	return us
}

type UAiMcpServer struct {
	ctx           context.Context
	lCtx          context.Context
	lCancel       context.CancelFunc
	name          string
	version       string
	log           *zap.SugaredLogger
	opt           *uaimcp.Option
	server        *server.MCPServer
	streamableSrv *server.StreamableHTTPServer
	sseSrv        *server.SSEServer
}

func (u *UAiMcpServer) shutdown() error {
	switch u.opt.Type {
	case uaimcp.MCP_Type_stdio:
		{
			return errors.New("stdio server not support")
		}
	case uaimcp.MCP_Type_http_sse:
		{
			err := u.sseSrv.Shutdown(u.ctx)
			if err != nil {
				return err
			}
			break
		}
	case uaimcp.MCP_Type_streamable_http:
		{
			err := u.streamableSrv.Shutdown(u.ctx)
			if err != nil {
				return err
			}
			break
		}
	default:
		{
			return errors.New(fmt.Sprintf("%s not support", u.opt.Type))
		}
	}
	return nil
}

func (u *UAiMcpServer) Start(ctx context.Context) (done <-chan struct{}, err error) {
	doneC := make(chan struct{})
	u.lCtx, u.lCancel = context.WithCancel(ctx)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				u.log.Errorf("panic recovered in MCP SERVER %+v ListenAndServe: %v", u.name, r)
				u.log.Errorf("stack: %s", utils.StackSkip(2, -1))
			}
			close(doneC)
		}()
		switch u.opt.Type {
		case uaimcp.MCP_Type_stdio:
			{
				err = errors.New("stdio server not support")
				break
			}
		case uaimcp.MCP_Type_http_sse:
			{
				err := u.sseSrv.Start(u.opt.BaseURL)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					panic(err)
				}
				break
			}
		case uaimcp.MCP_Type_streamable_http:
			{
				err := u.streamableSrv.Start(u.opt.BaseURL)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					panic(err)
				}
				break
			}
		default:
			err = errors.New(fmt.Sprintf("%s not support", u.opt.Type))
		}
	}()
	go func() {
		<-u.lCtx.Done()
		err = u.shutdown()
		if err != nil {
			return
		}
	}()

	return doneC, err
}

func (u *UAiMcpServer) Stop() error {
	u.lCancel()
	return nil
}

func (u *UAiMcpServer) RestStart() (done <-chan struct{}, err error) {
	err = u.Stop()
	if err != nil {
		return nil, err
	}
	return u.Start(u.ctx)
}
func (u *UAiMcpServer) NewTool(name string, opts ...mcp.ToolOption) mcp.Tool {
	return mcp.NewTool(name, opts...)
}

func (u *UAiMcpServer) AddTool(tool mcp.Tool, handler server.ToolHandlerFunc) {
	u.server.AddTool(tool, handler)
}

// Stdio 用于启动一个标准输入输出服务器,注意，该服务必须独占主服务进程，独占stdio的输入输出，不可以使用协程运行，否则，会出现不预测错误
func (u *UAiMcpServer) Stdio() error {
	// 调用server.ServeStdio函数，传入u.server参数
	return server.ServeStdio(u.server)
}

// WithRecovery adds a middleware that recovers from panics in tool handlers.
func (u *UAiMcpServer) WithRecovery() server.ServerOption {
	return server.WithToolHandlerMiddleware(func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf(
						"panic recovered in %s tool handler: %v",
						request.Params.Name,
						r,
					)
					u.log.Errorf("panic recovered in %+v tool handler: %v", request, r)
					u.log.Errorf("stack: %s", utils.StackSkip(2, -1))
				}
			}()
			return next(ctx, request)
		}
	})
}

func (u *UAiMcpServer) newHooks() *server.Hooks {
	hooks := &server.Hooks{}

	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		fmt.Printf("beforeAny: %s, %v, %v\n", method, id, message)
	})
	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		fmt.Printf("onSuccess: %s, %v, %v, %v\n", method, id, message, result)
	})
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		fmt.Printf("onError: %s, %v, %v, %v\n", method, id, message, err)
	})
	hooks.AddBeforeInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest) {
		fmt.Printf("beforeInitialize: %v, %v\n", id, message)
	})
	hooks.AddOnRequestInitialization(func(ctx context.Context, id any, message any) error {
		fmt.Printf("AddOnRequestInitialization: %v, %v\n", id, message)
		return nil
	})
	hooks.AddAfterInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest, result *mcp.InitializeResult) {
		fmt.Printf("afterInitialize: %v, %v, %v\n", id, message, result)
	})
	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		fmt.Printf("afterCallTool: %v, %v, %v\n", id, message, result)
	})
	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		fmt.Printf("beforeCallTool: %v, %v\n", id, message)
	})
	return hooks
}
