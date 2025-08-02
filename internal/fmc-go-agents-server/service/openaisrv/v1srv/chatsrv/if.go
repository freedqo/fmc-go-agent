package chatsrv

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/openaim/v1m/chatm"
)

type If interface {
	Completions(ctx context.Context, sessionId string, req chatm.ChatCompletionsReq) (res interface{}, err error)
}
