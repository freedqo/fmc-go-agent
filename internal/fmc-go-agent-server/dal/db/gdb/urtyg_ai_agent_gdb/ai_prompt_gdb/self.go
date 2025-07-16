package ai_prompt_gdb

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal/db/dbif/urtyg_ai_agent_if/ai_prompt_if"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/dalm/dbm/urtyg_ai_agent/query"
	"gorm.io/gorm"
)

func newSelfIF(gdb *gorm.DB, genQ *query.Query) ai_prompt_if.SelfIf {
	return &SelfIF{
		db:   gdb,
		genQ: genQ,
	}
}

type SelfIF struct {
	db   *gorm.DB
	genQ *query.Query
}

var _ ai_prompt_if.SelfIf = &SelfIF{}
