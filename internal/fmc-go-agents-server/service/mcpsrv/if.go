package mcpsrv

import (
	"context"
	"github.com/cloudwego/eino/components/tool"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp/faimcpclient/mcp2eino"
	"net/http"
)

type If interface {
	ServerIf
	ClientIf
}

type ServerIf interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
type ClientIf interface {
	GetEinoTools(ctx context.Context) ([]tool.BaseTool, error)
	mcp2eino.If
}
