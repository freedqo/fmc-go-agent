package uaivectordb

import (
	"context"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"time"
)

func newEmbedding(ctx context.Context, opt *Option) (eb embedding.Embedder, err error) {
	// TODO Modify component configuration here.
	config := &openai.EmbeddingConfig{
		BaseURL: opt.IRVModel.BaseURL + "/v1",
		APIKey:  opt.IRVModel.APIKey,
		Model:   opt.IRVModel.Model,
		Timeout: time.Duration(opt.IRVModel.Timeout) * time.Second,
	}
	eb, err = openai.NewEmbedder(ctx, config)
	if err != nil {
		return nil, err
	}
	return eb, nil
}
