package knowledge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/extm/Knowledgem"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/extm/usercenterm"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/store/middleware"
	"github.com/freedqo/fmc-go-agents/pkg/httpclient"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"net/http"
	"net/url"
)

func New(opt *httpclient.Option) If {
	if opt == nil {
		return nil
	}
	if !opt.Enable {
		return nil
	}
	u := &UserCenter{
		opt: opt,
		c:   httpclient.NewHttpClient("Knowledge", opt),
	}
	// 注册日志中间件
	u.c.Use(middleware.HttpClientLoggerMiddleware(u.c))
	return u
}

type UserCenter struct {
	opt *httpclient.Option
	c   *httpclient.HttpClient
}

func (u *UserCenter) GetFontMCPTools(ctx context.Context, req Knowledgem.GetFontMCPToolsReq) (res *Knowledgem.GetFontMCPToolsResp, err error) {
	res = &Knowledgem.GetFontMCPToolsResp{}
	cReq := u.c.NewRequest(nil)
	cRes := u.c.NewResponse(res)
	err = u.c.Get(ctx, "/knowledge/v1/t-mcp-tool-info/getAll", cReq, cRes)
	if err != nil {
		return nil, err
	}
	if res.Code != 200 {
		return nil, errors.New(res.Msg)
	}
	return res, nil
}

func (u *UserCenter) Sm2Login(ctx context.Context, req usercenterm.Sm2LoginReq) (res *usercenterm.Sm2LoginResp, err error) {
	res = &usercenterm.Sm2LoginResp{}
	// 处理加密
	encryptData := ""
	var loginInfoBytes []byte
	if loginInfoBytes, err = json.Marshal(req); err != nil {
		return nil, errors.New(fmt.Sprint("序列化入参异常:", err))
	}
	//获取sm2加密公钥
	var publicKeyObj *usercenterm.Sm2EncodeType
	if publicKeyObj, err = u.getSm2EncodePublicKey(ctx); err != nil {
		return nil, err
	}
	//构造加密请求数据
	if encryptData, err = utils.Sm2Encrypt(loginInfoBytes, publicKeyObj.PublicKey); err != nil {
		return nil, err
	}

	cReq := u.c.NewRequest(url.Values{
		"encryptData": []string{encryptData},
	})
	cRes := u.c.NewResponse(res)
	err = u.c.Post(ctx, "/usercenter/login/sm2Login", cReq, cRes)
	if err != nil {
		return nil, err
	}
	if res.Code != 200 {
		return nil, errors.New(res.Msg)
	}
	return res, nil
}

func (u *UserCenter) InvalidToken(ctx context.Context, req usercenterm.InvalidTokenReq) (res *usercenterm.InvalidTokenResp, err error) {
	res = &usercenterm.InvalidTokenResp{}
	cReq := u.c.NewRequest(nil)
	cRes := u.c.NewResponse(res)
	cReq.Headers["Tokenid"] = req.TokenId
	err = u.c.Get(ctx, "/usercenter/v2/sysUser/invalid", cReq, cRes)
	if err != nil {
		return nil, err
	}
	if res.Code != 200 && res.Code != 5003 {
		return nil, errors.New(res.Msg)
	}
	if res.Code == 5003 {
		return nil, errors.New(res.Msg)
	}
	return res, nil
}
func (u *UserCenter) GetMenuList(ctx context.Context, req usercenterm.GetMenuListReq) (res *usercenterm.GetMenuListResp, err error) {
	res = &usercenterm.GetMenuListResp{}
	cReq := u.c.NewRequest(nil)
	cRes := u.c.NewResponse(res)
	cReq.Headers["Tokenid"] = req.TokenId
	err = u.c.Get(ctx, "/usercenter/v2/sysUser/getMenuList", cReq, cRes)
	if err != nil {
		return nil, err
	}
	if res.Code != 200 {
		return nil, errors.New(res.Msg)
	}
	return res, nil
}

// getSm2EncodePublicKey 获取sm2加密公钥
func (u *UserCenter) getSm2EncodePublicKey(ctx context.Context) (res *usercenterm.Sm2EncodeType, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()
	response := &usercenterm.Sm2EncodeType_Response{}
	cReq := u.c.NewRequest(nil)
	cReq.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	cRes := u.c.NewResponse(response)
	err = u.c.Get(ctx, "/usercenter/login/encodeType", cReq, cRes)
	if err != nil {
		return nil, err
	}
	if response.Code != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("获取sm2加密公钥失败：%d-%s", response.Code, response.Msg))
	}
	if response.Data.PublicKey == "" {
		return nil, errors.New("获取sm2加密公钥为空！")
	}
	return &response.Data, nil
}

var _ If = &UserCenter{}
