package dal

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/db"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/ext"
	ucharmodel2 "github.com/freedqo/fmc-go-agents/pkg/fai/faicharmodel"
)

func New(ctx context.Context, opt *config.Config) If {
	return &Dal{
		ucm: ucharmodel2.New(ctx, opt.UCM),
		db:  db.New(opt.Db),
		ext: ext.New(ctx, opt.Ext),
	}
}

type Dal struct {
	ucm ucharmodel2.If
	db  db.If
	ext ext.If
}

func (d *Dal) Ext() ext.If {
	return d.ext
}

func (d *Dal) Db() db.If {
	return d.db
}

func (d *Dal) Cm() ucharmodel2.If {
	return d.ucm
}

var _ If = &Dal{}
