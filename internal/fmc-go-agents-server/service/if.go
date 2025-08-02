package service

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/knowdbsrv"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/mcpsrv"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/openaisrv"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/promptsrv"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/sessionsrv"
	"github.com/freedqo/fmc-go-agents/pkg/commif"
)

type If interface {
	commif.MonitorIf
	OpenAi() openaisrv.If
	Session() sessionsrv.If
	KnowDb() knowdbsrv.If
	MCP() mcpsrv.If
	MidInvalidTokenIdByUserCenter(ctx context.Context, tokenId string) error
	Prompt() promptsrv.If
}
