package db

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal/db/dbif"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal/db/gdb"
	"github.com/freedqo/fmc-go-agent/pkg/ugrom"
)

// New 构建信息显示车辆段管理服务API数据接口
func New(opt *ugrom.Option) If {
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
