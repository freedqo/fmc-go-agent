package promptm

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/dbm"
	"time"
)

type QueryReq struct {
	// 提示词内容(md格式)
	Content *string `json:"Content" column:"content" form:"Content"`
	// 创建时间
	CreatedAt *string `json:"CreatedAt" column:"created_at" form:"CreatedAt"`
	// 提示词描述
	Description *string `json:"Description" column:"description" form:"Description"`
	// 提示词唯一ID
	ID *string `json:"ID" column:"id" form:"ID"`
	// 提示词名称
	Name *string `json:"Name" column:"name" form:"Name"`
	// 提示词类型
	Type *string `json:"Type" column:"type" form:"Type"`
	// 更新时间
	UpdatedAt *string `json:"UpdatedAt" column:"updated_at" form:"UpdatedAt"`
	// 创建用户ID
	UserID *string `json:"UserID" column:"user_id" form:"UserID"`
	// 排序字段，例如 "字段名 asc" 或 "字段名 desc"
	OrderBy *string `json:"OrderBy" form:"OrderBy"`
	// 是否模糊查询
	IsLike bool `json:"IsLike" form:"IsLike"`
	// 分页信息
	Page *dbm.Page `json:"Page" form:"Page"`
}

type QueryResp struct {
	List []QueryRespData `json:"list"`
	// 分页信息
	Page *dbm.Page `json:"Page" form:"Page"`
}
type QueryRespData struct {
	ID          string     `json:"id"`          // 提示词唯一ID
	UserID      string     `json:"userID"`      // 创建用户ID
	Type        string     `json:"type"`        // 提示词类型
	Name        string     `json:"name"`        // 提示词名称
	Description string     `json:"description"` // 提示词描述
	Content     string     `json:"content"`     // 提示词内容
	CreatedAt   *time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   *time.Time `json:"updated_at"`  // 更新时间
}
