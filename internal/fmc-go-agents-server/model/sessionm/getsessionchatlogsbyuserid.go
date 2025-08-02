package sessionm

type QuerySessionChatLogsByUserReq struct {
	UserId     string  `json:"userId"`     // 用户id
	PromptType *string `json:"promptType"` // 提示词类型
}

type QuerySessionChatLogsByUserResp struct {
	SessionId  string     `json:"sessionId"`  // 会话id
	PromptType string     `json:"promptType"` // 提示词类型
	ChatLogs   []*ChatLog `json:"chatLogs"`   // 聊天记录
}

type ChatLog struct {
	Role      string `json:"role"`      // 消息角色
	Content   string `json:"content"`   // 消息内容
	Timestamp int64  `json:"timestamp"` // 消息时间戳
}
