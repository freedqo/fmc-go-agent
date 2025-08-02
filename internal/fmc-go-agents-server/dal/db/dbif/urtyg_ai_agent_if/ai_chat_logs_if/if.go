package ai_chat_logs_if

import "context"

type If interface {
	Gen() GenIf
	Self() SelfIf
}

type SelfIf interface {
	GetMaxOrder(ctx context.Context, sessionID string) (int64, error)
}
