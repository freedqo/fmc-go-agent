package uaimcpclient

import (
	"context"
	"github.com/cloudwego/eino/components/tool"
	"github.com/freedqo/fmc-go-agent/pkg/commif"
)

type If interface {
	commif.MonitorIf
	Initialize(ctx context.Context) error
	DToEinoTools(ctx context.Context) []tool.BaseTool
}
