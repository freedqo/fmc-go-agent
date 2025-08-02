package faiagent

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/embedding"
	"strconv"

	redispkg "github.com/cloudwego/eino-examples/quickstart/eino_assistant/pkg/redis"
	"github.com/cloudwego/eino/schema"
	redisCli "github.com/redis/go-redis/v9"

	"github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/retriever"
)

// newRetriever component initialization function of node 'RedisRetriever' in graph 'FAiAgent'
func (u *FAiAgent) newRetriever(ctx context.Context, redisClient *redisCli.Client, eb embedding.Embedder) (rtr retriever.Retriever, err error) {
	// TODO Modify component configuration here.
	config := &redis.RetrieverConfig{
		Client:       redisClient,
		Index:        fmt.Sprintf("%s%s", redispkg.RedisPrefix, redispkg.IndexName),
		Dialect:      2,
		ReturnFields: []string{redispkg.ContentField, redispkg.MetadataField, redispkg.DistanceField},
		TopK:         8,
		VectorField:  redispkg.VectorField,
		DocumentConverter: func(ctx context.Context, doc redisCli.Document) (*schema.Document, error) {
			resp := &schema.Document{
				ID:       doc.ID,
				Content:  "",
				MetaData: map[string]any{},
			}
			for field, val := range doc.Fields {
				if field == redispkg.ContentField {
					resp.Content = val
				} else if field == redispkg.MetadataField {
					resp.MetaData[field] = val
				} else if field == redispkg.DistanceField {
					distance, err := strconv.ParseFloat(val, 64)
					if err != nil {
						continue
					}
					resp.WithScore(1 - distance)
				}
			}

			return resp, nil
		},
	}
	config.Embedding = eb
	rtr, err = redis.NewRetriever(ctx, config)
	if err != nil {
		return nil, err
	}
	return rtr, nil
}
