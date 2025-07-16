package uaivectordbx

import (
	"context"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaivectordb"
	"go.uber.org/zap"
)

var db uaivectordb.If

func NewUAiVectorDb(ctx context.Context, option *uaivectordb.Option, log *zap.SugaredLogger) uaivectordb.If {
	db = uaivectordb.New(ctx, option, log)
	return db
}

func UAiVectorDb() uaivectordb.If {
	if db == nil {
		panic("uaivectordb is nil")
	}
	return db
}
