package ext

import (
	"context"
	"errors"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/config"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal/ext/usercenter"
)

func New(ctx context.Context, opt *config.ExtOption) If {
	return &Ext{
		opt:        opt,
		userCenter: usercenter.New(opt.UserCenter),
	}
}

type Ext struct {
	opt        *config.ExtOption
	userCenter usercenter.If
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
