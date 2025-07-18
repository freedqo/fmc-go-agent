// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameAi_chat_logs = "ai_chat_logs"

// Ai_chat_logs mapped from table <ai_chat_logs>
type Ai_chat_logs struct {
	ID        string     `gorm:"column:id;type:varchar(60);primaryKey;comment:记录唯一id" json:"id"`                                                                                          // 记录唯一id
	UserID    string     `gorm:"column:user_id;type:varchar(50);not null;index:idx_user_id,priority:1;comment:用户ID" json:"user_id"`                                                       // 用户ID
	SessionID string     `gorm:"column:session_id;type:varchar(512);not null;uniqueIndex:uniq_session_order,priority:1;index:idx_session_id,priority:1;comment:会话唯一标识" json:"session_id"` // 会话唯一标识
	Role      string     `gorm:"column:role;type:enum('user','assistant','system');not null;comment:消息角色" json:"role"`                                                                    // 消息角色
	Content   string     `gorm:"column:content;type:text;not null;comment:消息内容" json:"content"`                                                                                           // 消息内容
	Order     int32      `gorm:"column:order;type:int unsigned;not null;uniqueIndex:uniq_session_order,priority:2;comment:消息顺序" json:"order"`                                             // 消息顺序
	CreatedAt *time.Time `gorm:"column:created_at;type:timestamp;index:idx_created_at,priority:1;default:CURRENT_TIMESTAMP;comment:消息创建时间" json:"created_at"`                             // 消息创建时间
}

// TableName Ai_chat_logs's table name
func (*Ai_chat_logs) TableName() string {
	return TableNameAi_chat_logs
}
