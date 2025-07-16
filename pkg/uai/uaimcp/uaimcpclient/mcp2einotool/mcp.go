package mcp2einotool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type Config struct {
	// Cli is the MCP (Model Control Protocol) client, ref: https://github.com/mark3labs/mcp-go?tab=readme-ov-file#tools
	// Notice: should Initialize with server before use
	Cli client.MCPClient
	// ToolNameList specifies which tools to fetch from MCP server
	// If empty, all available tools will be fetched
	ToolNameList []string
	// ToolCallResultHandler is a function that processes the result after a tool call completes
	// It can be used for custom processing of tool call results
	// If nil, no additional processing will be performed
	ToolCallResultHandler func(ctx context.Context, name string, result *mcp.CallToolResult) (*mcp.CallToolResult, error)
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

		ret = append(ret, &toolHelper{
			cli: conf.Cli,
			info: &schema.ToolInfo{
				Name:        t.Name,
				Desc:        t.Description,
				ParamsOneOf: schema.NewParamsOneOfByOpenAPIV3(inputSchema),
			},
			toolCallResultHandler: conf.ToolCallResultHandler,
		})
	}

	return ret, nil
}

type toolHelper struct {
	cli                   client.MCPClient
	info                  *schema.ToolInfo
	toolCallResultHandler func(ctx context.Context, name string, result *mcp.CallToolResult) (*mcp.CallToolResult, error)
}

func (m *toolHelper) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return m.info, nil
}

func (m *toolHelper) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	result, err := m.cli.CallTool(ctx, mcp.CallToolRequest{
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

	if m.toolCallResultHandler != nil {
		result, err = m.toolCallResultHandler(ctx, m.info.Name, result)
		if err != nil {
			return "", fmt.Errorf("failed to execute mcp tool call result handler: %w", err)
		}
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
func (m *toolHelper) StreamableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (*schema.StreamReader[string], error) {
	// 创建一个管道，用于接收结果
	r, w := schema.Pipe[string](2)
	// 控制超时
	lctx, c := context.WithTimeout(ctx, 120*time.Second)
	// 创建一个goroutine，用于调用工具
	go func() {
		defer w.Close()
		select {
		case <-lctx.Done():
			{
				c()
				return
			}
		default:
			{
				// 调用工具
				result, err := m.cli.CallTool(ctx, mcp.CallToolRequest{
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
				// 如果调用工具失败，则返回错误
				if err != nil {
					w.Send("", fmt.Errorf("failed to call mcp tool: %w", err))
					return
				}

				// 如果有工具调用结果处理器，则执行处理器
				if m.toolCallResultHandler != nil {
					result, err = m.toolCallResultHandler(ctx, m.info.Name, result)
					if err != nil {
						w.Send("", fmt.Errorf("failed to call mcp tool: %w", fmt.Errorf("failed to execute mcp tool call result handler: %w", err)))
						return
					}
				}
				// 将结果转换为字符串
				marshaledResult, err := sonic.MarshalString(result)
				if err != nil {
					w.Send("", fmt.Errorf("failed to marshal mcp tool result: %v", err))
					return
				}
				// 如果结果有错误，则返回错误
				if result.IsError {
					w.Send("", fmt.Errorf("failed to marshal mcp tool result: %v", marshaledResult))
					return
				}
				// 将结果发送到管道
				w.Send(marshaledResult, nil)
				return
			}
		}
	}()

	return r, nil
}
