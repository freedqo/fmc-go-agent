package urag

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	db "github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/db"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb"
	"github.com/freedqo/fmc-go-agents/pkg/fconfvp"
)

var dbi db.If
var urag fvectordb.If

func GetURag(ctx context.Context) (fvectordb.If, db.If) {
	if urag == nil || dbi == nil {
		opt := config.NewDefault()
		_, err := fconfvp.New(opt, config.AppName)
		if err != nil {
			panic(err)
		}
		db1 := db.New(opt.Db)
		dbif := fvectordb.NewDocDbIf(db1.Gdb().Urtyg_ai_agent().Ai_knowledge_documents().Self().UpdateDocumentsStatus,
			db1.Gdb().Urtyg_ai_agent().Ai_knowledge_chunks().Self().SaveChunksData)
		g, err := fvectordb.New(ctx, opt.UVector, dbif)
		if err != nil {
			panic(err)
		}
		dbi = db1
		urag = g
	}
	return urag, dbi
}
