package sessionm

import "github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/dalm/dbm/urtyg_ai_agent/model"

type QuerySessionChatLogsByUserReq struct {
	UserId     string  `json:"userId"`     // 用户id
	PromptType *string `json:"promptType"` // 提示词类型
}

type QuerySessionChatLogsByUserResp struct {
	SessionId  string                `json:"sessionId"`  // 会话id
	PromptType string                `json:"promptType"` // 提示词类型
	ChatLogs   []*model.Ai_chat_logs `json:"chatLogs"`   // 聊天记录
}
