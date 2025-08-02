package promptm

import "time"

type CreatReq struct {
	UserID      string `json:"userID"`      // 创建用户ID
	Type        string `json:"type"`        // 提示词类型
	Name        string `json:"name"`        // 提示词名称
	Description string `json:"description"` // 提示词描述
	Content     string `json:"content"`     // 提示词内容
}

type CreatResp struct {
	ID          string     `json:"id"`          // 提示词唯一ID
	UserID      string     `json:"userID"`      // 创建用户ID
	Type        string     `json:"type"`        // 提示词类型
	Name        string     `json:"name"`        // 提示词名称
	Description string     `json:"description"` // 提示词描述
	Content     string     `json:"content"`     // 提示词内容
	CreatedAt   *time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   *time.Time `json:"updated_at"`  // 更新时间
}
