package uaiagent

import (
	"context"
	"time"
)

// newLambda component initialization function of node 'InputToQuery' in graph 'UAiAgent'
func newLambda(ctx context.Context, input *UserMessage, opts ...any) (output string, err error) {
	return input.Query, nil
}

// newLambda2 component initialization function of node 'InputToHistory' in graph 'UAiAgent'
func (u *UAiAgent) newLambda2(ctx context.Context, input *UserMessage, opts ...any) (output map[string]any, err error) {
	return map[string]any{
		"sessionId": input.SessionId,
		"content":   input.Query,
		"history":   input.History,
		"date":      time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}
