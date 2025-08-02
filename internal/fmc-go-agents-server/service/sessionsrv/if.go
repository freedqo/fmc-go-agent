package sessionsrv

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/sessionm"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faiagent/mem"
)

type If interface {
	// QuerySessionChatLogsByUser 根据用户和既有的会话信息,返回会话的id和聊天记录,ut-ai-font 特殊应用接口
	QuerySessionChatLogsByUser(ctx context.Context, req sessionm.QuerySessionChatLogsByUserReq) (*sessionm.QuerySessionChatLogsByUserResp, error)
	// CreatSession 创建一个会话（返回一个会话ID）
	CreatSession(ctx context.Context, req sessionm.CreatSessionReq) (*sessionm.CreatSessionResp, error)
	// UserSessionList 获取用户的会话列表（简化数据,所有或者分页）
	UserSessionList(ctx context.Context, req sessionm.UserSessionListReq) (*sessionm.UserSessionListResp, error)
	// SessionChatLogList 获取用户的会话列表（改对话的完整对话历史,所有）
	SessionChatLogList(ctx context.Context, req sessionm.SessionChatLogListReq) (*sessionm.SessionChatLogListResp, error)
	// DeleteSessions 删除多个对话
	DeleteSessions(ctx context.Context, req sessionm.DeleteSessionsReq) (*sessionm.DeleteSessionsResp, error)
	// DeleteChatLogs 删除多条对话内容
	DeleteChatLogs(ctx context.Context, req sessionm.DeleteChatLogsReq) (*sessionm.DeleteChatLogsResp, error)
	mem.MemoryIf
}
