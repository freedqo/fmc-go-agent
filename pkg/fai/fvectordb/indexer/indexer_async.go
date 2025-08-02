package indexer

import (
	"context"
	"fmt"
	common2 "github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb/common"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb/config"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/indexer/es8"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
)

// newAsyncIndexer component initialization function of node 'Indexer2' in graph 'rag'
func newAsyncIndexer(ctx context.Context, conf *config.Option) (idr indexer.Indexer, err error) {
	indexerConfig := &es8.IndexerConfig{
		Client:    conf.Client,
		Index:     conf.IndexName,
		BatchSize: 10,
		DocumentToFields: func(ctx context.Context, doc *schema.Document) (field2Value map[string]es8.FieldValue, err error) {
			var knowledgeName string
			if value, ok := ctx.Value(common2.KnowledgeName).(string); ok {
				knowledgeName = value
			} else {
				err = fmt.Errorf("必须提供知识库名称")
				return
			}
			if doc.MetaData != nil {
				// 存储ext数据
				marshal, _ := sonic.Marshal(getExtData(doc))
				doc.MetaData[common2.FieldExtra] = string(marshal)
			}
			return map[string]es8.FieldValue{
				common2.FieldContent: {
					Value:    doc.Content,
					EmbedKey: common2.FieldContentVector,
				},
				common2.FieldExtra: {
					Value: doc.MetaData[common2.FieldExtra],
				},
				common2.KnowledgeName: {
					Value: knowledgeName,
				},
				common2.FieldQAContent: {
					Value:    doc.MetaData[common2.FieldQAContent],
					EmbedKey: common2.FieldQAContentVector,
				},
			}, nil
		},
	}
	embeddingIns11, err := common2.NewEmbedding(ctx, conf)
	if err != nil {
		return nil, err
	}
	indexerConfig.Embedding = embeddingIns11
	idr, err = es8.NewIndexer(ctx, indexerConfig)
	if err != nil {
		return nil, err
	}
	return idr, nil
}
