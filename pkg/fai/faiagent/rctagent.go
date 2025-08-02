package faiagent

import (
	"context"
	"errors"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"io"
)

func (u *FAiAgent) newReactAgent(ctx context.Context, cm model.ToolCallingChatModel, tools []tool.BaseTool) (*react.Agent, error) {
	// 创建一个agent
	// TODO Modify component configuration here.

	config := &react.AgentConfig{
		ToolCallingModel: cm, // 将ChatModel实例赋值给config的Model字段
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: tools, // 将工具赋值给config的ToolsConfig.Tools字段
		},
		MessageModifier:       nil,
		MaxStep:               12,
		ToolReturnDirectly:    map[string]struct{}{},
		StreamToolCallChecker: u.firstChunkStreamToolCallChecker,
		GraphName:             "",
		ModelNodeName:         "",
		ToolsNodeName:         "",
	}
	// 创建一个Agent实例
	agent, err := react.NewAgent(ctx, config)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

// FilterToolsCall 函数用于过滤工具调用
func (u *FAiAgent) FilterToolsCall(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
	defer sr.Close()
	for {
		msg, err := sr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// finish
				break
			}
			return false, err
		}
		if msg.Content != "" {
			if u.wPeekToolsStreamOut != nil {
				u.wPeekToolsStreamOut.Send(msg, nil)
			}
			u.log.Infof("调用工具前时，有msg输出，影响流式判断应答: %s", msg.Content)
		}
		if len(msg.ToolCalls) > 0 {
			u.log.Infof("发现需要调用的工具: %v", msg.ToolCalls)
			return true, nil
		}
	}

	return false, nil
}
func (u *FAiAgent) firstChunkStreamToolCallChecker(_ context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
	defer sr.Close()

	for {
		msg, err := sr.Recv()
		if err == io.EOF {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if msg.Content != "" {
			u.log.Infof("调用工具前，大模型输出: %s", msg.Content)
		}
		if len(msg.ToolCalls) > 0 {
			u.log.Infof("发现需要调用的工具: %v", msg.ToolCalls)
			return true, nil
		}

		if len(msg.Content) == 0 { // skip empty chunks at the front
			continue
		}

		return false, nil
	}
}
