package service

import (
	"context"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/knowdbsrv"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/openaisrv"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/promptsrv"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/sessionsrv"
	"github.com/freedqo/fmc-go-agent/pkg/commif"
)

type If interface {
	commif.MonitorIf
	OpenAi() openaisrv.If
	Session() sessionsrv.If
	KnowDb() knowdbsrv.If
	MidInvalidTokenIdByUserCenter(ctx context.Context, tokenId string) error
	Prompt() promptsrv.If
}
