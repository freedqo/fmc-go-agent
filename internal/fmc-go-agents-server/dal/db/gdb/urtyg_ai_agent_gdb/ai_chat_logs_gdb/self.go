package ai_chat_logs_gdb

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/db/dbif/urtyg_ai_agent_if/ai_chat_logs_if"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/dbm/urtyg_ai_agent/query"
	"gorm.io/gorm"
)

func newSelfIF(gdb *gorm.DB, genQ *query.Query) ai_chat_logs_if.SelfIf {
	return &SelfIF{
		db:   gdb,
		genQ: genQ,
	}
}

type SelfIF struct {
	db   *gorm.DB
	genQ *query.Query
}

func (s *SelfIF) GetMaxOrder(ctx context.Context, sessionID string) (int64, error) {
	var maxOrder *int32 // 用指针接收，避免零值干扰（区分"无结果"和"结果为0"）
	err := s.genQ.Ai_chat_logs.
		WithContext(ctx).
		Where(s.genQ.Ai_chat_logs.SessionID.Eq(sessionID)).
		Select(s.genQ.Ai_chat_logs.Order.Max()).
		Scan(&maxOrder)
	if err != nil {
		return 0, err
	}
	if maxOrder == nil {
		return 0, nil
	}
	return int64(*maxOrder), nil
}

var _ ai_chat_logs_if.SelfIf = &SelfIF{}
