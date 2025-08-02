package ext

import (
	"context"
	"errors"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/ext/knowledge"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/ext/usercenter"
)

func New(ctx context.Context, opt *config.ExtOption) If {
	return &Ext{
		opt:        opt,
		userCenter: usercenter.New(opt.UserCenter),
		knowledge:  knowledge.New(opt.Knowledge),
	}
}

type Ext struct {
	opt        *config.ExtOption
	userCenter usercenter.If
	knowledge  knowledge.If
}

func (e *Ext) Knowledge() (knowledge.If, error) {
	if e.userCenter == nil {
		return nil, errors.New("knowledge is disable")
	}
	if e.opt.Knowledge == nil {
		return nil, errors.New("knowledge is disable")
	}
	if !e.opt.Knowledge.Enable {
		return nil, errors.New("knowledge is disable")
	}
	return e.knowledge, nil
}

func (e *Ext) UserCenter() (usercenter.If, error) {
	if e.userCenter == nil {
		return nil, errors.New("UserCenter is disable")
	}
	if e.opt.UserCenter == nil {
		return nil, errors.New("UserCenter is disable")
	}
	if !e.opt.UserCenter.Enable {
		return nil, errors.New("UserCenter is disable")
	}
	return e.userCenter, nil
}

var _ If = &Ext{}
