package config

import (
	"github.com/elastic/go-elasticsearch/v8"
)

type Option struct {
	Client    *elasticsearch.Client
	IndexName string // es index name
	// embedding 时使用
	APIKey         string
	BaseURL        string
	EmbeddingModel string
}

func (x *Option) Copy() *Option {
	return &Option{
		Client:    x.Client,
		IndexName: x.IndexName,
		// embedding 时使用
		APIKey:         x.APIKey,
		BaseURL:        x.BaseURL,
		EmbeddingModel: x.EmbeddingModel,
	}
}
