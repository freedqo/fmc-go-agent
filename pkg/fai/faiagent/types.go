package faiagent

import "github.com/cloudwego/eino/schema"

type UserMessage struct {
	SessionId string            `json:"sessionId"`
	PromptId  string            `json:"promptId"`
	Query     string            `json:"query"`
	History   []*schema.Message `json:"history"`
}
