package usercenter

import (
	"context"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/dalm/extm/usercenterm"
)

type If interface {
	Sm2Login(ctx context.Context, req usercenterm.Sm2LoginReq) (res *usercenterm.Sm2LoginResp, err error)
	GetMenuList(ctx context.Context, req usercenterm.GetMenuListReq) (res *usercenterm.GetMenuListResp, err error)
	InvalidToken(ctx context.Context, req usercenterm.InvalidTokenReq) (res *usercenterm.InvalidTokenResp, err error)
}
