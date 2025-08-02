package fvectordb

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb/common"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb/config"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb/indexer"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb/retriever"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
)

const (
	scoreThreshold = 1.05 // 设置一个很小的阈值
	esTopK         = 50
	esTryFindDoc   = 10
)

func New(ctx context.Context, opt *Option, docDbi DocDbIf) (If, error) {
	if opt == nil {
		return nil, errors.New("opt is nil")
	}
	if len(opt.IndexName) == 0 {
		return nil, fmt.Errorf("indexName is empty")
	}
	es8C, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: opt.Es8.Addresses,
		Username:  opt.Es8.Username,
		Password:  opt.Es8.Password,
	})
	conf := &config.Option{
		Client:         es8C,
		IndexName:      opt.IndexName,
		APIKey:         opt.Embedding.APIKey,
		BaseURL:        opt.Embedding.BaseURL,
		EmbeddingModel: opt.Embedding.Model,
	}

	// 确保es index存在
	err = common.CreateIndexIfNotExists(ctx, conf.Client, conf.IndexName)
	if err != nil {
		return nil, err
	}
	buildIndex, err := indexer.BuildIndexer(ctx, conf)
	if err != nil {
		return nil, err
	}
	buildIndexAsync, err := indexer.BuildIndexerAsync(ctx, conf, &openai.ChatModelConfig{
		Model:   opt.Qa.Model,
		BaseURL: opt.Qa.BaseURL,
		APIKey:  opt.Qa.APIKey,
	})
	if err != nil {
		return nil, err
	}
	buildRetriever, err := retriever.BuildRetriever(ctx, conf)
	if err != nil {
		return nil, err
	}
	qaCtx := context.WithValue(ctx, common.RetrieverFieldKey, common.FieldQAContentVector)
	qaRetriever, err := retriever.BuildRetriever(qaCtx, conf)
	if err != nil {
		return nil, err
	}

	return &FRag{
		opt:               opt,
		indexBuilder:      buildIndex,
		indexBuilderAsync: buildIndexAsync,
		retriever:         buildRetriever,
		qaRetriever:       qaRetriever,
		client:            conf.Client,
		conf:              conf,
		docDbi:            docDbi,
	}, nil
}

type FRag struct {
	opt               *Option                                        // 配置参数
	indexBuilder      compose.Runnable[any, []string]                // 同步构建索引器
	indexBuilderAsync compose.Runnable[[]*schema.Document, []string] // 异步构建索引器
	retriever         compose.Runnable[string, []*schema.Document]   // 同步检索器
	qaRetriever       compose.Runnable[string, []*schema.Document]   // 异步检索器
	client            *elasticsearch.Client                          // es8 client
	conf              *config.Option                                 // 配置
	docDbi            DocDbIf                                        // 文档数据库读写接口
}

func (x *FRag) Embedder() embedding.Embedder {
	//TODO implement me
	panic("implement me")
}

// GetKnowledgeBaseList 获取知识库列表
func (x *FRag) GetKnowledgeBaseList(ctx context.Context) (list []string, err error) {
	names := "distinct_knowledge_names"
	query := search.NewRequest()
	query.Size = utils.Of(0) // 不返回原始文档
	query.Aggregations = map[string]types.Aggregations{
		names: {
			Terms: &types.TermsAggregation{
				Field: utils.Of(common.KnowledgeName),
				Size:  utils.Of(10000),
			},
		},
	}
	res, err := search.NewSearchFunc(x.client)().
		Request(query).
		Do(ctx)
	if err != nil {
		return
	}
	if res.Aggregations == nil {
		return
	}
	termsAgg, ok := res.Aggregations[names].(*types.StringTermsAggregate)
	if !ok || termsAgg == nil {
		err = errors.New("failed to parse terms aggregation")
		return
	}
	for _, bucket := range termsAgg.Buckets.([]types.StringTermsBucket) {
		list = append(list, bucket.Key.(string))
	}
	return
}
