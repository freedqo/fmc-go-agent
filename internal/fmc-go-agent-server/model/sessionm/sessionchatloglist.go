package sessionm

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/dalm/dbm/urtyg_ai_agent/model"
)

type SessionChatLogListReq struct {
	SessionId string `json:"sessionId"`
}
type SessionChatLogListResp struct {
	SessionId string                `json:"sessionId"`
	ChatLogs  []*model.Ai_chat_logs `json:"chatLogs"`
}
