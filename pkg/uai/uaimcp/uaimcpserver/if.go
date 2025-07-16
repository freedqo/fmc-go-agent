package uaimcpserver

import (
	"github.com/freedqo/fmc-go-agent/pkg/commif"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type If interface {
	NewTool(name string, opts ...mcp.ToolOption) mcp.Tool
	AddTool(tool mcp.Tool, handler server.ToolHandlerFunc)
	commif.MonitorIf
}
