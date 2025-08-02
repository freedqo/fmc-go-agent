package dal

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/db"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/ext"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faicharmodel"
)

type If interface {
	Cm() faicharmodel.If
	Db() db.If
	Ext() ext.If
}
