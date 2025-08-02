package einosrv

import (
	"context"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faiagent"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb"
)

type If interface {
	VectorDb() fvectordb.If
	UAiAgent(ctx context.Context, sessionId string) (faiagent.If, error)
}
