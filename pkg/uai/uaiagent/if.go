package uaiagent

import (
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type If interface {
	AppendTools(name string, tool []tool.BaseTool) error
	Invoke(msg string) (*schema.Message, error)
	Stream(msg string) (*schema.StreamReader[*schema.Message], error)
	Collect(inMsg *schema.StreamReader[string]) (*schema.Message, error)
	Transform(inMsg *schema.StreamReader[string]) (*schema.StreamReader[*schema.Message], error)
}
