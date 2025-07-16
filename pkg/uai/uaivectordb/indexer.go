package uaivectordb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	redisCli "github.com/redis/go-redis/v9"
)

// newIndexer component initialization function of node 'RedisIndexer' in graph 'VectorDb'
func newIndexer(ctx context.Context, opt *Option, redisClient *redisCli.Client) (idr indexer.Indexer, err error) {
	
	config := &redis.IndexerConfig{
		Client:    redisClient,
		KeyPrefix: RedisPrefix,
		BatchSize: 1,
		DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*redis.Hashes, error) {
			if doc.ID == "" {
				doc.ID = uuid.New().String()
			}
			key := doc.ID

			metadataBytes, err := json.Marshal(doc.MetaData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal metadata: %w", err)
			}

			return &redis.Hashes{
				Key: key,
				Field2Value: map[string]redis.FieldValue{
					ContentField:  {Value: doc.Content, EmbedKey: VectorField},
					MetadataField: {Value: metadataBytes},
				},
			}, nil
		},
	}

	embeddingIns11, err := newEmbedding(ctx, opt)
	if err != nil {
		return nil, err
	}
	config.Embedding = embeddingIns11
	idr, err = redis.NewIndexer(ctx, config)
	if err != nil {
		return nil, err
	}
	return idr, nil
}
