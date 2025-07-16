package einosrv

import (
	"context"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaiagent"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaivectordb"
)

type If interface {
	VectorDb() uaivectordb.If
	UAiAgent(ctx context.Context, sessionId string) uaiagent.If
}
