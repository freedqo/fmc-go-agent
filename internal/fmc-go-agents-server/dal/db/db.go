package db

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/db/dbif"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/db/gdb"
	"github.com/freedqo/fmc-go-agents/pkg/fgrom"
)

// New 构建信息显示车辆段管理服务API数据接口
func New(opt *fgrom.Option) If {
	if opt == nil {
		panic("option is nil")
	}
	db := Db{
		gdb: gdb.New(opt),
	}

	return &db
}

type Db struct {
	gdb dbif.If
}

func (d Db) Gdb() dbif.If {
	return d.gdb
}
