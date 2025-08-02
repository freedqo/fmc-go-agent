package common

import (
	"context"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb/config"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
)

func NewEmbedding(ctx context.Context, conf *config.Option) (eb embedding.Embedder, err error) {
	econf := &openai.EmbeddingConfig{
		APIKey:     conf.APIKey,
		Model:      conf.EmbeddingModel,
		Dimensions: utils.Of(1024),
		Timeout:    0,
		BaseURL:    conf.BaseURL,
	}
	if econf.APIKey == "" {
		econf.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if econf.BaseURL == "" {
		econf.BaseURL = os.Getenv("OPENAI_BASE_URL")
	}
	if econf.Model == "" {
		econf.Model = "text-embedding-3-large"
	}
	eb, err = openai.NewEmbedder(ctx, econf)
	if err != nil {
		return nil, err
	}
	return eb, nil
}
