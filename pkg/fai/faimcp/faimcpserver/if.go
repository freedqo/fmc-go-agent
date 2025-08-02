package faimcpserver

import (
	"github.com/freedqo/fmc-go-agents/pkg/commif"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"net/http"
)

type If interface {
	NewTool(name string, opts ...mcp.ToolOption) mcp.Tool
	AddTool(tool mcp.Tool, handler server.ToolHandlerFunc)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	commif.MonitorIf
}
