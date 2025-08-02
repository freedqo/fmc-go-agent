package v1srv

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/openaisrv/v1srv/chatsrv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/pagination"
)

type If interface {
	Chat() chatsrv.If
	Models(ctx context.Context, opts ...option.RequestOption) (res *pagination.Page[openai.Model], err error)
}
