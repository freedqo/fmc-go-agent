package extmcp

import (
	"context"
	"github.com/cloudwego/eino/components/tool"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaimcp"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaimcp/uaimcpclient"
)

func New(ctx context.Context, mapMcp *map[string]*uaimcp.Option) If {
	mcp := make(map[string]uaimcpclient.If, 0)
	for k, v := range *mapMcp {
		mcp[k] = uaimcpclient.New(ctx, k, v)
	}
	return &ExtMCP{
		mcp: mcp,
	}
}

type ExtMCP struct {
	mcp map[string]uaimcpclient.If
}

func (e *ExtMCP) EinoTools(ctx context.Context) []tool.BaseTool {
	tools := make([]tool.BaseTool, 0)
	for _, v := range e.mcp {
		var v1 uaimcpclient.If
		v1 = v
		tools = append(tools, v1.DToEinoTools(ctx)...)
	}
	return tools
}

var _ If = &ExtMCP{}
