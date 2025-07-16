package dal

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal/db"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal/ext"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaicharmodel"
)

type If interface {
	Cm() uaicharmodel.If
	Db() db.If
	Ext() ext.If
}
