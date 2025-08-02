package ai_knowledge_chunks_gdb

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/db/dbif/urtyg_ai_agent_if/ai_knowledge_chunks_if"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/dbm/urtyg_ai_agent/model"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/dbm/urtyg_ai_agent/query"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb/entity"
	"github.com/gogf/gf/v2/frame/g"
	v1 "github.com/wangle201210/go-rag/server/api/rag/v1"
	"gorm.io/gorm"
)

func newSelfIF(gdb *gorm.DB, genQ *query.Query) ai_knowledge_chunks_if.SelfIf {
	return &SelfIF{
		db:   gdb,
		genQ: genQ,
	}
}

type SelfIF struct {
	db   *gorm.DB
	genQ *query.Query
}

var _ ai_knowledge_chunks_if.SelfIf = &SelfIF{}

// SaveChunksData 批量保存知识块数据
func (s *SelfIF) SaveChunksData(ctx context.Context, documentsId int64, chunks []entity.KnowledgeChunks) error {
	if len(chunks) == 0 {
		return nil
	}
	status := int(v1.StatusIndexing)
	data := make([]*model.Ai_knowledge_chunks, 0, len(chunks))
	for _, doc := range chunks {
		st := false
		if doc.Status == 1 {
			st = true
		}
		v := model.Ai_knowledge_chunks{
			ID:             doc.Id,
			KnowledgeDocID: doc.KnowledgeDocId,
			ChunkID:        doc.ChunkId,
			Content:        doc.Content,
			Ext:            doc.Ext,
			Status:         &st,
			CreatedAt:      doc.CreatedAt,
			UpdatedAt:      doc.UpdatedAt,
		}
		data = append(data, &v)

	}
	err := s.genQ.Ai_knowledge_chunks.WithContext(ctx).Save(data...)
	if err != nil {
		g.Log().Errorf(ctx, "SaveChunksData err=%+v", err)
		status = int(v1.StatusFailed)
	}
	_, err = s.genQ.Ai_knowledge_documents.WithContext(ctx).
		Where(s.genQ.Ai_knowledge_documents.ID.Eq(documentsId)).
		UpdateColumns(map[string]interface{}{
			"status": status,
		})
	if err != nil {
		return err
	}
	return err
}
