// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package ai_session_logs_if

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/dalm/dbm/urtyg_ai_agent/model"
)

type GenIf interface {
	// Add 新增
	Add(data ...*model.Ai_session_logs) error
	// Upt 更新
	Upt(data ...*model.Ai_session_logs) error
	// Save 保存
	Save(data ...*model.Ai_session_logs) error
	// Del 删除
	Del(data ...*model.Ai_session_logs) error
	// First 查询第一条
	First(query *model.Ai_session_logs) (data *model.Ai_session_logs,err error)
	// Find 通用查询
	Find(query *model.Ai_session_logs_QueryReq) (data []*model.Ai_session_logs,total int64,err error)
}

