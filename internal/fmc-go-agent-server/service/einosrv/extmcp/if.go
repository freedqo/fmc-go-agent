package extmcp

import (
	"context"
	"github.com/cloudwego/eino/components/tool"
)

type If interface {
	EinoTools(ctx context.Context) []tool.BaseTool
}
