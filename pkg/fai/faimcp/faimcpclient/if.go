package faimcpclient

import (
	"context"
	"github.com/cloudwego/eino/components/tool"
	"github.com/freedqo/fmc-go-agents/pkg/commif"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp/faimcpclient/mcp2eino"
	"github.com/mark3labs/mcp-go/mcp"
)

type If interface {
	commif.MonitorIf
	Initialize(ctx context.Context) error
	ServerInfo() *mcp.InitializeResult
	SubToolMidFunc(fun *mcp2eino.If)
	UnSubToolMidFunc(fun *mcp2eino.If)
	DToEinoTools(ctx context.Context) []tool.BaseTool
	CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}
