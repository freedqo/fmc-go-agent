package ai_knowledge_documents_gdb

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/db/dbif/urtyg_ai_agent_if/ai_knowledge_documents_if"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/dbm/urtyg_ai_agent/query"
	"gorm.io/gorm"
)

func newSelfIF(gdb *gorm.DB, genQ *query.Query) ai_knowledge_documents_if.SelfIf {
	return &SelfIF{
		db:   gdb,
		genQ: genQ,
	}
}

type SelfIF struct {
	db   *gorm.DB
	genQ *query.Query
}

var _ ai_knowledge_documents_if.SelfIf = &SelfIF{}

// UpdateDocumentsStatus 更新文档状态
func (s *SelfIF) UpdateDocumentsStatus(ctx context.Context, documentsId int64, status int) error {
	data := map[string]interface{}{
		"status": status,
	}
	_, err := s.genQ.Ai_knowledge_documents.WithContext(ctx).
		Where(s.genQ.Ai_knowledge_documents.ID.Eq(documentsId)).
		Updates(data)
	if err != nil {
		return err
	}
	return err
}
