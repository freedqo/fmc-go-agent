package uaivectordb

import (
	"context"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/compose"
	"github.com/redis/go-redis/v9"
)

type If interface {
	Redis() *redis.Client
	Embedder() embedding.Embedder
	Runnable() compose.Runnable[document.Source, []string]
	BuildDir(ctx context.Context, dir string) (err error)
	BuildFile(ctx context.Context, filepath string) (err error)
}
