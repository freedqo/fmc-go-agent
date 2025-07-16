package service

import (
	"context"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/config"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/dalm/extm/usercenterm"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/einosrv"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/knowdbsrv"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/mcpsrv"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/msgsrv"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/openaisrv"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/promptsrv"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/sessionsrv"
	"sync"
)

func New(ctx context.Context, opt *config.Config) If {
	// 实例根服务
	s := &Service{
		ctx: ctx, // 上下文
		opt: opt, // 配置
	}
	// 实例消息服务，提供内置独立（端口）的WebSocket Server支持
	s.msg = msgsrv.New(opt)

	// 实例数据访问层
	s.dal = dal.New(ctx, opt)

	// 实例mcp服务，提供内置独立（端口）的MCP Server支持
	s.mcp = mcpsrv.New(ctx, opt, s.dal.Ext(), s.msg.Publish, s.msg.WaitingResMsg)

	// 实例session服务
	s.session = sessionsrv.New(ctx, opt, s.dal)

	// 实例knowdb服务
	s.knowdb = knowdbsrv.New()

	// 实例prompt服务
	s.prompt = promptsrv.New(opt, s.dal)

	// 实例eino服务,并注入会话管理
	s.eino = einosrv.New(ctx, opt, s.dal, s.session)

	// 实例openai服务
	s.openai = openaisrv.New(ctx, opt, s.dal, s.eino, s.msg)

	return s
}

type Service struct {
	ctx     context.Context
	lCtx    context.Context
	lCancel context.CancelFunc
	opt     *config.Config
	dal     dal.If
	openai  openaisrv.If
	eino    einosrv.If
	session sessionsrv.If
	mcp     mcpsrv.If
	msg     msgsrv.If
	knowdb  knowdbsrv.If
	prompt  promptsrv.If
}

func (s *Service) Prompt() promptsrv.If {
	return s.prompt
}

func (s *Service) KnowDb() knowdbsrv.If {
	return s.knowdb
}

func (s *Service) Start(ctx context.Context) (done <-chan struct{}, err error) {
	donec := make(chan struct{})
	once := sync.Once{}
	s.lCtx, s.lCancel = context.WithCancel(ctx)

	msgdone, err := s.msg.Start(s.lCtx)
	if err != nil {
		return nil, err
	}

	go func() {
		<-msgdone
		once.Do(func() {
			close(donec)
		})
	}()

	return donec, nil
}

func (s *Service) Stop() error {
	s.lCancel()
	return nil
}

func (s *Service) RestStart() (done <-chan struct{}, err error) {
	err = s.Stop()
	if err != nil {
		return nil, err
	}
	return s.Start(s.ctx)
}

func (s *Service) MidInvalidTokenIdByUserCenter(ctx context.Context, tokenId string) error {
	userCenter, err := s.dal.Ext().UserCenter()
	if err != nil {
		return err
	}
	_, err = userCenter.InvalidToken(ctx, usercenterm.InvalidTokenReq{TokenId: tokenId})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Session() sessionsrv.If {
	return s.session
}

// OpenAi 方法返回 Service 结构体的 openai 字段
func (s *Service) OpenAi() openaisrv.If {
	// 返回 openai 字段
	return s.openai
}

var _ If = &Service{}
