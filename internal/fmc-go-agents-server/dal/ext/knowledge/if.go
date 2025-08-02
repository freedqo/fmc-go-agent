package knowledge

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/extm/Knowledgem"
)

type If interface {
	GetFontMCPTools(ctx context.Context, req Knowledgem.GetFontMCPToolsReq) (res *Knowledgem.GetFontMCPToolsResp, err error)
}
