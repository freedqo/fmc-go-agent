package fvectordb

import (
	"context"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb/entity"
)

type If interface {
	retriever.Retriever
	Embedder() embedding.Embedder
	GetKnowledgeBaseList(ctx context.Context) (list []string, err error)
	Index(ctx context.Context, req *IndexReq) (ids []string, err error)
	IndexAsync(ctx context.Context, req *IndexAsyncReq) (ids []string, err error)
	DeleteDocument(ctx context.Context, documentID string) error
}
type DocDbIf interface {
	// UpdateDocumentsStatus 更新文档构建状态
	UpdateDocumentsStatus(ctx context.Context, documentsId int64, status int) error
	// SaveChunksData 保存文档切片
	SaveChunksData(ctx context.Context, documentsId int64, chunks []entity.KnowledgeChunks) error
}
