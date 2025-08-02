package faicharmodel

import (
	"github.com/cloudwego/eino/components/model"
	"github.com/openai/openai-go"
)

type ModelProvider string

type If interface {
	//BindTools(tools []*schema.ToolInfo) error
	// ToolCallingChatModel eino cm 接口
	UChatModelIf
	// SetProvider 设置模型提供方
	SetProvider(provider ModelProvider) error
	V1() *openai.Client
}

type UChatModelIf interface {
	model.ToolCallingChatModel
	model.ChatModel
}
