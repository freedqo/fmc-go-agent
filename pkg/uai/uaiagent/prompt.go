package uaiagent

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

var baseP = `
## 用户信息
- 不能暴露这些具体的用户信息
- sessionId: {sessionId},备注,这个数据你不能暴露在对话中

## 上下文信息
- Current Date: {date}
- Related Documents: |-
==== doc start ====
  {documents}
==== doc end ====`

type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

// newChatTemplate component initialization function of node 'ChatTemplate' in graph 'UAiAgent'
func (u *UAiAgent) newChatTemplate(ctx context.Context, sessionId string) (ctp prompt.ChatTemplate, err error) {
	// TODO Modify component configuration here.
	men := u.memory.GetConversation(sessionId, true)
	config := &ChatTemplateConfig{
		FormatType: schema.FString,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(men.GetPrompt() + baseP),
			schema.MessagesPlaceholder("history", true),
			schema.UserMessage("{content}"),
		},
	}
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}
