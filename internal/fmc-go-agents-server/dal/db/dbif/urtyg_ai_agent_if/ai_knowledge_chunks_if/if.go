package ai_knowledge_chunks_if

import (
	"context"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb/entity"
)

type If interface {
	Gen() GenIf
	Self() SelfIf
}

type SelfIf interface {
	SaveChunksData(ctx context.Context, documentsId int64, chunks []entity.KnowledgeChunks) error
}
