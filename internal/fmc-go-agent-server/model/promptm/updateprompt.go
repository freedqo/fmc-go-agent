package promptm

import "time"

type UpdateReq struct {
	ID          string `json:"id"`          // 模板唯一ID
	UserID      string `json:"userID"`      // 创建用户ID
	Type        string `json:"type"`        // 模板类型
	Name        string `json:"name"`        // 模板名称
	Description string `json:"description"` // 模板描述
	Content     string `json:"content"`     // 模板内容
}
type UpdateResp struct {
	ID          string     `json:"id"`          // 模板唯一ID
	UserID      string     `json:"userID"`      // 创建用户ID
	Type        string     `json:"type"`        // 模板类型
	Name        string     `json:"name"`        // 模板名称
	Description string     `json:"description"` // 模板描述
	Content     string     `json:"content"`     // 模板内容
	CreatedAt   *time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   *time.Time `json:"updated_at"`  // 更新时间
}
