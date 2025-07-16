package promptsrv

import (
	"context"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/promptm"
)

type If interface {
	GetPromptTemplate(ctx context.Context, req struct{}) (*promptm.GetPromptTemplateResp, error)
	// Creat 添加提示词模板,添加用户自定义的模板
	Creat(ctx context.Context, req promptm.CreatReq) (*promptm.CreatResp, error)
	// Delete 删除提示词模板,删除用户自定义的模板
	Delete(ctx context.Context, req promptm.DeleteReq) (*promptm.DeleteResp, error)
	// ModifySessionPrompt 修改会话提示词模板,修改指定的会话使用的提示词模板
	ModifySessionPrompt(ctx context.Context, req promptm.ModifySessionPromptReq) (*promptm.ModifySessionPromptResp, error)
	// Query 查询用户可用的提示词模板,固化的模板、其他用户共享的模板、用户自定义的提示词模板
	Query(ctx context.Context, req promptm.QueryReq) (*promptm.QueryResp, error)
	// Update 修改用户自定义提示词模板,修改用户自定义的模板
	Update(ctx context.Context, req promptm.UpdateReq) (*promptm.UpdateResp, error)
}
