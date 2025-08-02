package chatsrv

import (
	"context"
	"errors"
	"github.com/cloudwego/eino/schema"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/openaim/v1m/chatm"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/einosrv"
)

// New函数用于创建一个新的Service实例
func New(ctx context.Context, opt *config.Config, dal dal.If, eino einosrv.If) If {
	// 返回一个新的Service实例，传入的参数包括上下文、配置和数据库访问层
	return &Service{
		ctx:  ctx,
		opt:  opt,
		dal:  dal,
		eino: eino,
	}
}

type Service struct {
	ctx  context.Context
	opt  *config.Config
	dal  dal.If
	eino einosrv.If
}

func (s *Service) Completions(ctx context.Context, sessionId string, req chatm.ChatCompletionsReq) (res interface{}, err error) {
	msg := make([]*schema.Message, 0)
	if req.Messages != nil {
		for _, v := range req.Messages {
			m := &schema.Message{
				Role:         schema.RoleType(v.Role),
				Content:      v.Content,
				MultiContent: nil,
				Name:         "",
				ToolCalls:    nil,
				ToolCallID:   "",
				ToolName:     "",
				ResponseMeta: nil,
				Extra:        nil,
			}
			msg = append(msg, m)
		}
	}
	if len(msg) == 0 {
		return nil, errors.New("messages is empty")
	}
	ag, err := s.eino.UAiAgent(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	if req.Stream != nil {
		if *req.Stream {
			// TODO: 实现流式处理,在控制器回写
			stream, err := ag.Stream(msg[len(msg)-1].Content)
			if err != nil {
				return nil, err
			}
			return stream, nil
		} else {
			// TODO: 实现非流式处理，在控制器识别与返回
			generate, err := ag.Invoke(msg[len(msg)-1].Content)
			if err != nil {
				return nil, err
			}
			return generate, nil
		}
	} else {
		// TODO: 实现非流式处理
		generate, err := ag.Invoke(msg[len(msg)-1].Content)
		if err != nil {
			return nil, err
		}
		return generate, nil
	}
}

var _ If = &Service{}
