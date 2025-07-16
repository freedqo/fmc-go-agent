package v1srv

import (
	"context"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/config"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/iconsts"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/einosrv"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/msgsrv"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/openaisrv/v1srv/chatsrv"
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/pagination"
)

// New函数用于创建一个新的Service实例
func New(ctx context.Context, opt *config.Config, dal dal.If, eino einosrv.If, msg msgsrv.If) If {
	// 返回一个新的Service实例，传入的参数包括上下文、配置和数据库访问层
	return &Service{
		ctx:  ctx,
		opt:  opt,
		dal:  dal,
		chat: chatsrv.New(ctx, opt, dal, eino),
		eino: eino,
		msg:  msg,
	}
}

type Service struct {
	ctx  context.Context
	opt  *config.Config
	dal  dal.If
	chat chatsrv.If
	eino einosrv.If
	msg  msgsrv.If
}

func (s *Service) Chat() chatsrv.If {
	return s.chat
}

// Models 方法用于获取 openai.Model 的分页数据
// Models 函数用于获取 openai.Model 类型的分页数据
func (s *Service) Models(ctx context.Context, opts ...option.RequestOption) (res *pagination.Page[openai.Model], err error) {
	// 调用 s.dal.Cm().V1().Models.List 函数，传入 ctx 和 opts 参数，返回分页数据 res 和错误信息 err
	msg := &umsg.Message{
		MessageBase: umsg.MessageBase{
			ClientID:        "342432432423",
			Operate:         iconsts.Topic_Mqtt_Page_Redirection_Within_The_System_Tool_Ca,
			IsReplyOperate:  true,
			OperateID:       "342432432423",
			OperateDataType: "strint",
		},
		OperateData: "route",
	}
	uMsg := umsg.NewUMsg(msg, "McpServer", []string{umsg.ToWsServer, umsg.ToMqtt})
	s.msg.Publish(uMsg)
	return s.dal.Cm().V1().Models.List(ctx, opts...)
}

var _ If = &Service{}
