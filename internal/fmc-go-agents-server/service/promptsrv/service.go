package promptsrv

import (
	"context"
	"errors"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/iconsts"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/dbm/urtyg_ai_agent/model"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/promptm"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"time"
)

// New 函数用于创建一个新的Service实例
func New(opt *config.Config, dal dal.If) If {
	// 返回一个新的Service实例，传入的参数包括上下文、配置和数据库访问层
	s := &Service{
		opt: opt,
		dal: dal,
	}
	return s
}

type Service struct {
	opt *config.Config
	dal dal.If
}

func (s *Service) GetPromptTemplate(ctx context.Context, req struct{}) (*promptm.GetPromptTemplateResp, error) {
	res := promptm.GetPromptTemplateResp{List: make([]promptm.GetPromptTemplateRespData, 0)}
	for _, v := range iconsts.PromptDist {
		res.List = append(res.List, promptm.GetPromptTemplateRespData{
			Name:        v.Name,
			Description: v.Description,
			Content:     v.Content,
		})
	}
	return &res, nil
}

func (s *Service) Creat(ctx context.Context, req promptm.CreatReq) (*promptm.CreatResp, error) {
	now := time.Now()
	data := &model.Ai_prompt{
		ID:          utils.GetStringID(),
		UserID:      req.UserID,
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		IsShared:    false,
		SharedAt:    time.Now(),
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}
	// 提示词类型同名校验
	_, i, err := s.dal.Db().Gdb().Urtyg_ai_agent().Ai_prompt().Gen().Find(&model.Ai_prompt_QueryReq{
		Type: &req.Type,
	})
	if err != nil {
		return nil, err
	}
	if i > 0 {
		return nil, errors.New("提示词类型禁止重复")
	}
	// 提示词名称同名校验
	_, i, err = s.dal.Db().Gdb().Urtyg_ai_agent().Ai_prompt().Gen().Find(&model.Ai_prompt_QueryReq{
		Name: &req.Name,
	})
	if err != nil {
		return nil, err
	}
	if i > 0 {
		return nil, errors.New("提示词名称禁止重复")
	}
	// 约束判断
	_, i, err = s.dal.Db().Gdb().Urtyg_ai_agent().Ai_prompt().Gen().Find(&model.Ai_prompt_QueryReq{
		Type:   &req.Type,
		UserID: &req.UserID,
	})
	if err != nil {
		return nil, err
	}
	if i > 2 {
		return nil, errors.New("一个提示词类型,一个用户只能创建1个")
	}
	err = s.dal.Db().Gdb().Urtyg_ai_agent().Ai_prompt().Gen().Add(data)
	if err != nil {
		return nil, err
	}
	return &promptm.CreatResp{
		ID:          data.ID,
		UserID:      data.UserID,
		Type:        data.Type,
		Name:        data.Name,
		Description: data.Description,
		Content:     data.Content,
		CreatedAt:   data.CreatedAt,
		UpdatedAt:   data.UpdatedAt,
	}, nil
}

func (s *Service) Delete(ctx context.Context, req promptm.DeleteReq) (*promptm.DeleteResp, error) {
	result, err := s.dal.Db().Gdb().Urtyg_ai_agent().GenQ().
		WithContext(ctx).Ai_prompt.
		Where(s.dal.Db().Gdb().Urtyg_ai_agent().GenQ().Ai_prompt.ID.In(req.Ids...)).
		Delete()
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return nil, nil
}

func (s *Service) ModifySessionPrompt(ctx context.Context, req promptm.ModifySessionPromptReq) (*promptm.ModifySessionPromptResp, error) {
	return nil, nil
}

func (s *Service) Query(ctx context.Context, req promptm.QueryReq) (*promptm.QueryResp, error) {
	list, _, err := s.dal.Db().Gdb().Urtyg_ai_agent().Ai_prompt().Gen().Find(&model.Ai_prompt_QueryReq{
		Content:     req.Content,
		CreatedAt:   req.CreatedAt,
		Description: req.Description,
		ID:          req.ID,
		IsShared:    nil,
		Name:        req.Name,
		SharedAt:    nil,
		Type:        req.Name,
		UpdatedAt:   req.UpdatedAt,
		UserID:      req.UserID,
		OrderBy:     req.OrderBy,
		IsLike:      req.IsLike,
		Page:        req.Page,
	})
	if err != nil {
		return nil, err
	}
	res := promptm.QueryResp{
		List: make([]promptm.QueryRespData, 0),
		Page: req.Page,
	}
	for _, v := range list {
		v1 := *v
		res.List = append(res.List, promptm.QueryRespData{
			ID:          v1.ID,
			UserID:      v1.UserID,
			Type:        v1.Type,
			Name:        v1.Name,
			Description: v1.Description,
			Content:     v1.Content,
			CreatedAt:   v1.CreatedAt,
			UpdatedAt:   v1.UpdatedAt,
		})
	}
	return &res, nil
}

func (s *Service) Update(ctx context.Context, req promptm.UpdateReq) (*promptm.UpdateResp, error) {
	first, err := s.dal.Db().Gdb().Urtyg_ai_agent().Ai_prompt().Gen().First(&model.Ai_prompt{ID: req.ID})
	if err != nil {
		return nil, err
	}
	if first == nil {
		return nil, errors.New("提示词不存在")
	}
	if first.Type != req.Type {
		return nil, errors.New("提示词类型禁止修改")
	}
	if first.Name != req.Name {
		_, i, err := s.dal.Db().Gdb().Urtyg_ai_agent().Ai_prompt().Gen().Find(&model.Ai_prompt_QueryReq{
			Name: &req.Name,
		})
		if err != nil {
			return nil, err
		}
		if i > 0 {
			return nil, errors.New("提示词名称禁止重复")
		}
	}
	now := time.Now()
	data := &model.Ai_prompt{
		ID:          first.ID,
		UserID:      first.UserID,
		Type:        first.Type,
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		IsShared:    false,
		SharedAt:    first.SharedAt,
		CreatedAt:   first.CreatedAt,
		UpdatedAt:   &now,
	}
	err = s.dal.Db().Gdb().Urtyg_ai_agent().Ai_prompt().Gen().Upt(data)
	if err != nil {
		return nil, err
	}
	return &promptm.UpdateResp{
		ID:          data.ID,
		UserID:      data.UserID,
		Type:        data.Type,
		Name:        data.Name,
		Description: data.Description,
		Content:     data.Content,
		CreatedAt:   data.CreatedAt,
		UpdatedAt:   data.UpdatedAt,
	}, nil
}

func (s *Service) loadBaseDate() {
	// 加载基础数据
	for _, v := range iconsts.PromptDist {
		v1 := model.Ai_prompt{
			ID:          v.ID,
			UserID:      "1", // 固化为一个不可编辑的人员
			Type:        v.Type,
			Name:        v.Name,
			Description: v.Description,
			Content:     v.Content,
			IsShared:    v.IsShared,
			SharedAt:    v.SharedAt,
			CreatedAt:   &v.SharedAt,
			UpdatedAt:   &v.SharedAt,
		}
		err := s.dal.Db().Gdb().Urtyg_ai_agent().Ai_prompt().Gen().Save(&v1)
		if err != nil {
			panic(err)
		}
	}
}
