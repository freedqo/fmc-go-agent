package uaiagent

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaiagent/mem"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"io"
	"sync"
)

// New 函数用于创建一个新的UAiAgent实例
func New(ctx context.Context, log *zap.SugaredLogger, cm model.ToolCallingChatModel, eb embedding.Embedder, redis *redis.Client, memdb mem.MemoryIf, sessionId string, tools []tool.BaseTool) If {
	// 创建一个新的UAiAgent实例
	uag := &UAiAgent{
		ctx:       ctx,
		sessionId: sessionId,
		mu:        sync.RWMutex{},
		// 初始化cbLog
		cbLog: newCbLog(log),
		// 初始化memory
		memory: mem.New(nil), // 先用默认配置
		log:    log,
	}
	if memdb != nil {
		uag.memory = memdb
	}
	var err error

	// 创建一个新的ChatTemplate实例
	uag.ctp, err = uag.newChatTemplate(ctx, sessionId)
	if err != nil {
		// 如果创建ChatTemplate实例失败，则抛出异常
		panic(err)
	}
	// 创建一个eino提供的服务本地工具集合
	einoLocalTools, err := uag.NetEinoLocalTools(ctx)
	if err != nil {
		panic(err)
	}

	tols := make([]tool.BaseTool, 0)
	tols = append(tols, einoLocalTools...)
	// 提供一个外部的工具集
	tols = append(tols, tools...)

	// 创建一个agent
	agent, err := uag.newReactAgent(ctx, cm, tols)
	if err != nil {
		return nil
	}

	// 创建一个新的Lambda实例
	uag.lba, err = uag.newLambda1(agent)
	if err != nil {
		// 如果创建Lambda实例失败，则抛出异常
		panic(err)
	}

	// 创建一个新的Retriever实例
	uag.rtr, err = uag.newRetriever(ctx, redis, eb)
	if err != nil {
		// 如果创建Retriever实例失败，则抛出异常
		panic(err)
	}

	// 构建运行图
	uag.g, uag.r, err = uag.buildRunnable(ctx, uag.ctp, uag.lba, uag.rtr)
	if err != nil {
		// 如果构建运行图失败，则抛出异常
		panic(err)
	}

	// 返回UAiAgent实例
	return uag
}

type UAiAgent struct {
	ctx                 context.Context                                 // 上下文
	sessionId           string                                          // 会话ID
	mu                  sync.RWMutex                                    // 互斥锁
	cbLog               callbacks.Handler                               // 日志回调
	g                   *compose.Graph[*UserMessage, *schema.Message]   // 代理运行图
	r                   compose.Runnable[*UserMessage, *schema.Message] // 代理运行图
	memory              mem.MemoryIf                                    // 对话缓存
	ctp                 prompt.ChatTemplate                             // 系统提示词模板
	lba                 *compose.Lambda                                 // Lambda代理
	rtr                 retriever.Retriever                             // 检索器
	log                 *zap.SugaredLogger                              // 日志
	wPeekToolsStreamOut *schema.StreamWriter[*schema.Message]           // 用来接收peek的流输出

}

func (u *UAiAgent) AppendTools(name string, tool []tool.BaseTool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	// 创建工具节点
	toolsNode, _ := compose.NewToolNode(u.ctx, &compose.ToolsNodeConfig{
		Tools:                tool,
		UnknownToolsHandler:  nil,
		ExecuteSequentially:  false,
		ToolArgumentsHandler: nil,
	})
	err := u.g.AddToolsNode("MCPTools", toolsNode)
	if err != nil {
		return err
	}
	return nil
}

// Invoke 函数用于运行一个代理，接受一个上下文、一个ID和一个消息作为参数，返回一个Message和一个错误
func (u *UAiAgent) Invoke(msg string) (*schema.Message, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	// 获取对话
	conversation := u.memory.GetConversation(u.sessionId, true)
	// 创建用户消息
	userMessage := &UserMessage{
		SessionId: u.sessionId,
		Query:     msg,
		History:   conversation.GetMessages(),
	}
	// 运行代理
	sr, err := u.r.Invoke(u.ctx, userMessage, compose.WithCallbacks(u.cbLog))
	if err != nil {
		return nil, fmt.Errorf("failed to invoke: %w", err)
	}
	// add user input to history
	conversation.Append(schema.UserMessage(msg))
	// add agent response to history
	conversation.Append(sr)

	return sr, nil
}

// Stream 函数用于运行一个代理，接受一个上下文、一个ID和一个消息作为参数，返回一个StreamReader和一个错误
func (u *UAiAgent) Stream(msg string) (*schema.StreamReader[*schema.Message], error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	// 获取对话
	conversation := u.memory.GetConversation(u.sessionId, true)
	// 创建用户消息
	userMessage := &UserMessage{
		SessionId: u.sessionId,
		Query:     msg,
		History:   conversation.GetMessages(),
	}
	// 运行代理
	sr, err := u.r.Stream(u.ctx, userMessage, compose.WithCallbacks(u.cbLog))
	if err != nil {
		return nil, err
	}
	// 启动一个goroutine来保存到内存中
	srs := sr.Copy(2)
	go func() {
		// for save to memory
		fullMsgs := make([]*schema.Message, 0)
		defer func() {
			// close stream if you used it
			srs[1].Close()

			// add user input to history
			conversation.Append(schema.UserMessage(msg))

			fullMsg, err := schema.ConcatMessages(fullMsgs)
			if err != nil {
				fmt.Println("error concatenating messages: ", err.Error())
			}
			// add agent response to history
			conversation.Append(fullMsg)
		}()
		for {
			select {
			case <-u.ctx.Done():
				fmt.Println("context done", u.ctx.Err())
				return
			default:
				chunk, err := srs[1].Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return
					}
				}

				fullMsgs = append(fullMsgs, chunk)
			}
		}
	}()
	return srs[0], nil
}

// Collect 函数用于运行一个代理，接受一个上下文、一个ID和一个消息作为参数，返回一个Message和一个错误
func (u *UAiAgent) Collect(inMsg *schema.StreamReader[string]) (*schema.Message, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	in, sw := schema.Pipe[*UserMessage](10)
	defer func() {
		in.Close()
		sw.Close()
	}()
	msg := ""
	// 获取对话
	conversation := u.memory.GetConversation(u.sessionId, true)
	go func() {
		for {
			select {
			case <-u.ctx.Done():
				{
					return
				}
			default:
				{
					msg, err := inMsg.Recv()
					if err != nil && errors.Is(err, io.EOF) {
						return
					}
					if err != nil {
						return
					}
					userMsg := &UserMessage{
						SessionId: u.sessionId,
						Query:     msg,
						History:   conversation.GetMessages(),
					}
					msg += msg
					isClose := sw.Send(userMsg, nil)
					if isClose {
						return
					}
				}
			}
		}
	}()
	// 运行代理
	sr, err := u.r.Collect(u.ctx, in, compose.WithCallbacks(u.cbLog))
	if err != nil {
		return nil, fmt.Errorf("failed to invoke: %w", err)
	}

	// add user input to history
	conversation.Append(schema.UserMessage(msg))
	// add agent response to history
	conversation.Append(sr)

	return sr, nil
}

// Transform 函数用于运行一个代理，接受一个上下文、一个ID和一个消息作为参数，返回一个Message和一个错误
func (u *UAiAgent) Transform(inMsg *schema.StreamReader[string]) (*schema.StreamReader[*schema.Message], error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	in, sw := schema.Pipe[*UserMessage](10)
	defer func() {
		in.Close()
		sw.Close()
	}()
	msgs := ""
	// 获取对话
	conversation := u.memory.GetConversation(u.sessionId, true)
	go func() {
		for {
			select {
			case <-u.ctx.Done():
				{
					return
				}
			default:
				{
					msg, err := inMsg.Recv()
					if err != nil && errors.Is(err, io.EOF) {
						return
					}
					if err != nil {
						return
					}
					userMsg := &UserMessage{
						SessionId: u.sessionId,
						Query:     msg,
						History:   conversation.GetMessages(),
					}
					msgs += msg
					isClose := sw.Send(userMsg, nil)
					if isClose {
						return
					}
				}
			}
		}
	}()
	// 运行代理
	sr, err := u.r.Transform(u.ctx, in, compose.WithCallbacks(u.cbLog))
	if err != nil {
		return nil, fmt.Errorf("failed to invoke: %w", err)
	}
	// 启动一个goroutine来保存到内存中
	srs := sr.Copy(2)
	go func() {
		// for save to memory
		fullMsgs := make([]*schema.Message, 0)

		defer func() {
			// close stream if you used it
			srs[1].Close()

			// add user input to history
			conversation.Append(schema.UserMessage(msgs))

			fullMsg, err := schema.ConcatMessages(fullMsgs)
			if err != nil {
				fmt.Println("error concatenating messages: ", err.Error())
			}
			// add agent response to history
			conversation.Append(fullMsg)
		}()
		for {
			select {
			case <-u.ctx.Done():
				fmt.Println("context done", u.ctx.Err())
				return
			default:
				chunk, err := srs[1].Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return
					}
				}

				fullMsgs = append(fullMsgs, chunk)

			}
		}
	}()
	return srs[0], nil
}

var _ If = &UAiAgent{}
