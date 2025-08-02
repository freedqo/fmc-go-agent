package v1srv

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/iconsts"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/msgm"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/einosrv"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/msgsrv"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/openaisrv/v1srv/chatsrv"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
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
	msg := &msgm.TAiAgentMessage{
		SessionId: "342432432423",
		Topic:     iconsts.Topic_Mqtt_To_Font_Ca,
		RespType:  1,
		OperateID: "342432432423",
		Data:      "route",
	}
	uMsg := fmsg.NewUMsg("OdqTest", msg, "McpServer", []string{fmsg.ToWsServer, fmsg.ToMqtt}, nil)
	s.msg.Publish(uMsg)
	return s.dal.Cm().V1().Models.List(ctx, opts...)
}

var _ If = &Service{}
