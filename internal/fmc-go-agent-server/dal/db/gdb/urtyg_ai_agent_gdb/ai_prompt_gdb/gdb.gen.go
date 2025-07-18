// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package ai_prompt_gdb

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal/db/dbif/urtyg_ai_agent_if/ai_prompt_if"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/dalm/dbm/urtyg_ai_agent/query"
	"gorm.io/gorm"
)

func New(gdb *gorm.DB, genQ *query.Query) ai_prompt_if.If {
	return &Ai_prompt_gdb{
		genIf:  newGenIF(gdb),
		selfIf: newSelfIF(gdb, genQ),
	}
}

type Ai_prompt_gdb struct {
	genIf  ai_prompt_if.GenIf
	selfIf ai_prompt_if.SelfIf
}

func (c *Ai_prompt_gdb) Gen() ai_prompt_if.GenIf {
	return c.genIf
}

func (c *Ai_prompt_gdb) Self() ai_prompt_if.SelfIf {
	return c.selfIf
}
