package openaisrv

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/einosrv"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/msgsrv"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/openaisrv/v1srv"
)

// New 函数用于创建一个新的Service实例
func New(ctx context.Context, opt *config.Config, dal dal.If, eino einosrv.If, msg msgsrv.If) If {
	// 返回一个Service实例，包含传入的context、config和dal
	return &Service{
		ctx:  ctx,
		opt:  opt,
		dal:  dal,
		v1:   v1srv.New(ctx, opt, dal, eino, msg), // 创建一个新的v1srv实例
		eino: eino,                                // 创建一个新的einosrv实例
	}
}

type Service struct {
	ctx  context.Context
	opt  *config.Config
	dal  dal.If
	v1   v1srv.If
	eino einosrv.If
}

func (s *Service) V1() v1srv.If {
	return s.v1
}

var _ If = &Service{}
