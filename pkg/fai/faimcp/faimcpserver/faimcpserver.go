package faimcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
	"net/http"
)

func New(ctx context.Context, opt *faimcp.Option, log *zap.SugaredLogger, name, version string) If {
	if name == "" {
		name = "ut-ai-mcp-server:" + utils.GetStringID()
	}
	if version == "" {
		version = "1.0.0"
	}
	us := &FAiMcpServer{
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

type FAiMcpServer struct {
	ctx           context.Context
	lCtx          context.Context
	lCancel       context.CancelFunc
	name          string
	version       string
	log           *zap.SugaredLogger
	opt           *faimcp.Option
	server        *server.MCPServer
	streamableSrv *server.StreamableHTTPServer
	sseSrv        *server.SSEServer
}

func (u *FAiMcpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch u.opt.Type {
	case faimcp.MCP_Type_stdio:
		{
			panic("stdio server not support")
		}
	case faimcp.MCP_Type_http_sse:
		{
			u.sseSrv.ServeHTTP(w, r)
			break
		}
	case faimcp.MCP_Type_streamable_http:
		{
			u.streamableSrv.ServeHTTP(w, r)
			break
		}
	default:
		{
			panic(fmt.Sprintf("%s not support", u.opt.Type))
		}
	}
	return
}
func (u *FAiMcpServer) shutdown() error {
	switch u.opt.Type {
	case faimcp.MCP_Type_stdio:
		{
			return errors.New("stdio server not support")
		}
	case faimcp.MCP_Type_http_sse:
		{
			err := u.sseSrv.Shutdown(u.ctx)
			if err != nil {
				return err
			}
			break
		}
	case faimcp.MCP_Type_streamable_http:
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

func (u *FAiMcpServer) Start(ctx context.Context) (done <-chan struct{}, err error) {
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
		case faimcp.MCP_Type_stdio:
			{
				err = errors.New("stdio server not support")
				break
			}
		case faimcp.MCP_Type_http_sse:
			{
				err := u.sseSrv.Start(u.opt.BaseURL)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					panic(err)
				}
				break
			}
		case faimcp.MCP_Type_streamable_http:
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

func (u *FAiMcpServer) Stop() error {
	u.lCancel()
	return nil
}

func (u *FAiMcpServer) RestStart() (done <-chan struct{}, err error) {
	err = u.Stop()
	if err != nil {
		return nil, err
	}
	return u.Start(u.ctx)
}
func (u *FAiMcpServer) NewTool(name string, opts ...mcp.ToolOption) mcp.Tool {
	return mcp.NewTool(name, opts...)
}

func (u *FAiMcpServer) AddTool(tool mcp.Tool, handler server.ToolHandlerFunc) {
	u.server.AddTool(tool, handler)
}

// Stdio 用于启动一个标准输入输出服务器,注意，该服务必须独占主服务进程，独占stdio的输入输出，不可以使用协程运行，否则，会出现不预测错误
func (u *FAiMcpServer) Stdio() error {
	// 调用server.ServeStdio函数，传入u.server参数
	return server.ServeStdio(u.server)
}

// WithRecovery adds a middleware that recovers from panics in tool handlers.
func (u *FAiMcpServer) WithRecovery() server.ServerOption {
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

// 辅助函数：将any类型转换为格式化的JSON字符串（处理序列化错误）
func anyToJSON(v any) string {
	if v == nil {
		return "null" // 空值直接返回"null"
	}
	// 使用MarshalIndent生成带缩进的JSON，便于阅读
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "[序列化失败] " + err.Error() // 处理序列化错误
	}
	return string(data)
}

func (u *FAiMcpServer) newHooks() *server.Hooks {
	hooks := &server.Hooks{}

	// 1. 通用前置：所有MCP请求进入后的第一个处理步骤
	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		u.log.Infof("MCP工具执行顺序 -> [通用前置处理] 开始处理任何MCP请求，方法：%s，请求ID：%s，请求内容：%s",
			method,
			anyToJSON(id),      // id转换为JSON
			anyToJSON(message)) // message转换为JSON
	})

	// 2. 初始化前置：初始化请求的前置准备
	hooks.AddBeforeInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest) {
		u.log.Infof("MCP工具执行顺序 -> [初始化前置] 准备处理初始化请求（如会话建立），请求ID：%s，初始化参数：%s",
			anyToJSON(id),
			anyToJSON(message)) // 指针类型也能被JSON序列化
	})

	// 3. 初始化处理中：执行初始化核心逻辑
	hooks.AddOnRequestInitialization(func(ctx context.Context, id any, message any) error {
		u.log.Infof("MCP工具执行顺序 -> [初始化处理中] 正在处理初始化请求，请求ID：%s，处理内容：%s",
			anyToJSON(id),
			anyToJSON(message))
		return nil
	})

	// 4. 初始化完成后：初始化请求处理完毕
	hooks.AddAfterInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest, result *mcp.InitializeResult) {
		u.log.Infof("MCP工具执行顺序 -> [初始化完成后] 初始化请求处理完毕，请求ID：%s，初始化参数：%s，返回结果：%s",
			anyToJSON(id),
			anyToJSON(message),
			anyToJSON(result))
	})

	// 5. 工具调用前置：调用工具前的准备
	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		u.log.Infof("MCP工具执行顺序 -> [工具调用前置] 准备调用工具，请求ID：%s，工具调用参数：%s",
			anyToJSON(id),
			anyToJSON(message))
	})

	// 6. 工具调用完成后：工具执行完毕
	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		u.log.Infof("MCP工具执行顺序 -> [工具调用完成后] 工具调用结束，请求ID：%s，调用参数：%s，工具返回结果：%s",
			anyToJSON(id),
			anyToJSON(message),
			anyToJSON(result))
	})

	// 7. 处理成功：整个请求处理完成且成功
	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		u.log.Infof("MCP工具执行顺序 -> [处理成功] 整个MCP请求处理完成且成功，方法：%s，请求ID：%s，请求内容：%s，处理结果：%s",
			method,
			anyToJSON(id),
			anyToJSON(message),
			anyToJSON(result))
	})

	// 8. 处理失败：整个请求处理过程中发生错误
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		u.log.Errorf("MCP工具执行 -> [处理失败] 整个MCP请求处理过程中发生错误，方法：%s，请求ID：%s，请求内容：%s，错误信息：%v",
			method,
			anyToJSON(id),
			anyToJSON(message),
			err) // err本身是error类型，直接打印即可
	})

	return hooks
}
