package mem

import "github.com/cloudwego/eino/schema"

type MemoryIf interface {
	GetConversation(id string, createIfNotExist bool) ConversationIf
}

type ConversationIf interface {
	Append(msg *schema.Message)
	GetMessages() []*schema.Message
	GetPrompt() string
}
