package ai_mcp_gdb

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/db/dbif/urtyg_ai_agent_if/ai_mcp_if"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/dbm/urtyg_ai_agent/query"
	"gorm.io/gorm"
)

func newSelfIF(gdb *gorm.DB, genQ *query.Query) ai_mcp_if.SelfIf {
	return &SelfIF{
		db:   gdb,
		genQ: genQ,
	}
}

type SelfIF struct {
	db   *gorm.DB
	genQ *query.Query
}

var _ ai_mcp_if.SelfIf = &SelfIF{}
