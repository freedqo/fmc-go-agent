package mcp2eino

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"time"
)

type Config struct {
	Cli          client.MCPClient
	ToolNameList []string
	MidFunc      map[*If]If
}

func GetTools(ctx context.Context, conf *Config) ([]tool.BaseTool, error) {
	listResults, err := conf.Cli.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list mcp tools fail: %w", err)
	}

	nameSet := make(map[string]struct{})
	for _, name := range conf.ToolNameList {
		nameSet[name] = struct{}{}
	}

	ret := make([]tool.BaseTool, 0, len(listResults.Tools))
	for _, t := range listResults.Tools {
		if len(conf.ToolNameList) > 0 {
			if _, ok := nameSet[t.Name]; !ok {
				continue
			}
		}
		marshaledInputSchema, err := sonic.Marshal(t.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("conv mcp tool input schema fail(marshal): %w, tool name: %s", err, t.Name)
		}
		inputSchema := &openapi3.Schema{}
		err = sonic.Unmarshal(marshaledInputSchema, inputSchema)
		if err != nil {
			return nil, fmt.Errorf("conv mcp tool input schema fail(unmarshal): %w, tool name: %s", err, t.Name)
		}
		Extra := make(map[string]any)
		js, err1 := json.Marshal(t.Annotations)
		if err1 == nil {
			_ = json.Unmarshal(js, &Extra)
		}
		ret = append(ret, &toolHelper{
			cli: conf.Cli,
			info: &schema.ToolInfo{
				Name:        t.Name,
				Desc:        t.Description,
				ParamsOneOf: schema.NewParamsOneOfByOpenAPIV3(inputSchema),
				Extra:       Extra,
			},
			midFunc: conf.MidFunc,
		})
	}
	return ret, nil
}

type toolHelper struct {
	cli     client.MCPClient
	info    *schema.ToolInfo
	midFunc map[*If]If
}

func (m *toolHelper) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return m.info, nil
}

func (m *toolHelper) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	data := make(map[string]interface{})
	err := sonic.Unmarshal([]byte(argumentsInJSON), &data)
	if err != nil {
		return "", fmt.Errorf("解析工具入参失败: %w", err)
	}
	sessionId, ok := data["sessionId"].(string)
	if !ok {
		return "", fmt.Errorf("工具入参:会话IDsessionId为空")
	}
	// 创建一个工具执行的中间件信息对象
	toolAcInfo := &EinoMcpMidInfo{
		SessionId: sessionId,
		Err:       nil,
		ToolActionInfo: ToolActionInfo{
			Id:        uuid.NewString(),
			Name:      m.info.Name,
			Des:       m.info.Desc,
			Args:      argumentsInJSON,
			Result:    nil,
			Message:   "",
			Status:    0,
			TimeStamp: time.Now().Unix(),
		},
	}
	// 获取中文别称
	title, ok := m.info.Extra["title"].(string)
	if ok {
		toolAcInfo.ToolActionInfo.Name = title
	}
	// 在调用工具前，回调一下钩子函数
	if m.midFunc != nil {
		for _, v := range m.midFunc {
			v.Before(ctx, toolAcInfo)
		}
	}
	var result *mcp.CallToolResult
	defer func() {
		var resErr error
		if err != nil {
			resErr = err
		}
		if result.IsError {
			resErr = fmt.Errorf("failed to call mcp tool, mcp server return error: %s", result.Result)
		}
		if result == nil {
			resErr = fmt.Errorf("failed to call mcp tool, mcp server return error: %s", "result is nil")
		}
		toolAcInfo.Err = resErr
		toolAcInfo.Result = result
		if m.midFunc != nil {
			for _, v := range m.midFunc {
				v.After(ctx, toolAcInfo)
			}
		}
	}()
	result, err = m.cli.CallTool(ctx, mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      m.info.Name,
			Arguments: json.RawMessage(argumentsInJSON),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to call mcp tool: %w", err)
	}
	marshaledResult, err := sonic.MarshalString(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal mcp tool result: %w", err)
	}
	if result.IsError {
		return "", fmt.Errorf("failed to call mcp tool, mcp server return error: %s", marshaledResult)
	}
	return marshaledResult, nil
}

//func (m *toolHelper) StreamableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (*schema.StreamReader[string], error) {
//	// 创建一个管道，用于接收结果
//	r, w := schema.Pipe[string](2)
//	// 控制超时
//	lctx, c := context.WithTimeout(ctx, 120*time.Second)
//	// 创建一个goroutine，用于调用工具
//	go func() {
//		defer w.Close()
//		select {
//		case <-lctx.Done():
//			{
//				c()
//				return
//			}
//		default:
//			{
//				// 创建一个工具执行的中间件信息对象
//				toolAcInfo := &EinoMcpMidInfo{
//					SessionId: "",
//					ToolActionInfo: ToolActionInfo{
//						Id:        uuid.NewString(),
//						Name:      m.info.Name,
//						Des:       m.info.Desc,
//						Args:      argumentsInJSON,
//						Result:    nil,
//						Message:   "",
//						Status:    0,
//						TimeStamp: time.Now().Unix(),
//					},
//				}
//				// 获取中文别称
//				title, ok := m.info.Extra["title"].(string)
//				if ok {
//					toolAcInfo.ToolActionInfo.Name = title
//				}
//				// 在调用工具前，回调一下钩子函数
//				if m.midFunc != nil {
//					for _, v := range m.midFunc {
//						v.Before(ctx, toolAcInfo)
//					}
//				}
//				// 调用工具
//				result, err := m.cli.CallTool(ctx, mcp.CallToolRequest{
//					Request: mcp.Request{
//						Method: "tools/call",
//					},
//					Params: struct {
//						Name      string    `json:"name"`
//						Arguments any       `json:"arguments,omitempty"`
//						Meta      *mcp.Meta `json:"_meta,omitempty"`
//					}{
//						Name:      m.info.Name,
//						Arguments: json.RawMessage(argumentsInJSON),
//					},
//				})
//				// 如果调用工具失败，则返回错误
//				if err != nil {
//					w.Send("", fmt.Errorf("failed to call mcp tool: %w", err))
//					return
//				}
//				// 将结果转换为字符串
//				marshaledResult, err := sonic.MarshalString(result)
//				if err != nil {
//					w.Send("", fmt.Errorf("failed to marshal mcp tool result: %v", err))
//					return
//				}
//				toolAcInfo.ToolActionInfo.Result = result
//				if m.midFunc != nil {
//					for _, v := range m.midFunc {
//						err = v.After(ctx, toolAcInfo)
//						if err != nil {
//							w.Send("", fmt.Errorf("failed to call mcp tool: %w", fmt.Errorf("failed to execute mcp tool call result handler: %w", err)))
//						}
//					}
//				}
//				// 如果结果有错误，则返回错误
//				if result.IsError {
//					w.Send("", fmt.Errorf("failed to marshal mcp tool result: %v", marshaledResult))
//					return
//				}
//				// 将结果发送到管道
//				w.Send(marshaledResult, nil)
//				return
//			}
//		}
//	}()
//
//	return r, nil
//}
