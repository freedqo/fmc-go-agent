package sessionsrv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudwego/eino/schema"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/iconsts"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/dbm/urtyg_ai_agent/model"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/sessionm"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/sessionsrv/dbconversation"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faiagent/mem"
	"github.com/freedqo/fmc-go-agents/pkg/fjwt"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"sort"
	"sync"
	"time"
)

// New 函数用于创建一个新的Service实例
func New(ctx context.Context, opt *config.Config, dal dal.If) If {
	// 返回一个新的Service实例，传入的参数包括上下文、配置和数据库访问层
	s := &Service{
		ctx:  ctx,
		opt:  opt,
		dal:  dal,
		uJwt: fjwt.New("utAiAgent", "aB#20150306"),
		mu:   sync.Mutex{},
	}
	return s
}

type Service struct {
	ctx  context.Context
	opt  *config.Config
	dal  dal.If
	uJwt fjwt.If
	mu   sync.Mutex
}

func (s *Service) GetConversation(sessionId string, createIfNotExist bool) mem.ConversationIf {
	s.mu.Lock()
	defer s.mu.Unlock()
	con := dbconversation.New(sessionId, s.uJwt, s.dal.Db().Gdb())
	return con
}

func (s *Service) UserSessionList(ctx context.Context, req sessionm.UserSessionListReq) (*sessionm.UserSessionListResp, error) {
	if req.UserId == "" {
		return nil, fmt.Errorf("userId is empty")
	}
	logs, i, err := s.dal.Db().Gdb().Urtyg_ai_agent().Ai_chat_logs().Gen().Find(&model.Ai_chat_logs_QueryReq{
		Content:   nil,
		CreatedAt: nil,
		ID:        nil,
		Order:     nil,
		Role:      nil,
		SessionID: nil,
		UserID:    &req.UserId,
		OrderBy:   nil,
		IsLike:    false,
		Page:      req.Page,
	})
	if err != nil {
		return nil, err
	}
	m1 := make(map[string][]*model.Ai_chat_logs, 0)
	// 按会话id的不同进行分类
	for _, v := range logs {
		_, ok := m1[v.SessionID]
		if !ok {
			m1[v.SessionID] = make([]*model.Ai_chat_logs, 0)
		}
		m1[v.SessionID] = append(m1[v.SessionID], v)

	}
	mapData := make(map[string]*sessionm.UserSessionListRespData, 0)

	// 对每个会话内的对话记录按order排序
	for _, v := range m1 {
		// 排序逻辑
		sort.Slice(v, func(i, j int) bool {
			// 2. 按 Order 升序
			return v[i].Order < v[j].Order
		})
		// 拿第一条记录的内容来做标题
		for _, v1 := range v {
			_, ok := mapData[v1.SessionID]
			if ok {
				break
			}
			msg := &schema.Message{}
			err := json.Unmarshal([]byte(v1.Content), msg)
			if err != nil {
				return nil, err
			}
			index := 20
			if len(msg.Content) <= index {
				index = len(msg.Content)
			}
			session := &sessionm.UserSessionListRespData{
				SessionId: v1.SessionID,
				Title:     msg.Content[:index],
				CreatAt:   v1.CreatedAt,
			}
			mapData[v1.SessionID] = session
		}
	}
	if req.Page != nil {
		req.Page.Total = int(i)
	}
	res := &sessionm.UserSessionListResp{
		SessionList: make([]*sessionm.UserSessionListRespData, 0),
		Page:        req.Page,
	}
	for _, v := range mapData {
		v1 := *v
		res.SessionList = append(res.SessionList, &v1)
	}
	sort.Slice(res.SessionList, func(i, j int) bool {
		// 获取两个元素的时间指针
		t1 := res.SessionList[i].CreatAt
		t2 := res.SessionList[j].CreatAt
		// 处理 nil 情况（nil 视为“最早”，排在最后）
		switch {
		case t1 == nil && t2 == nil:
			// 两者都为 nil，顺序不变
			return false
		case t1 == nil:
			// t1 为 nil，t2 非 nil → t1 排在 t2 后面
			return false
		case t2 == nil:
			// t2 为 nil，t1 非 nil → t1 排在 t2 前面
			return true
		default:
			// 两者都非 nil，比较时间：t1 比 t2 晚（更新），则 t1 排在前面
			return t1.After(*t2)
		}
	})
	return res, nil
}

func (s *Service) SessionChatLogList(ctx context.Context, req sessionm.SessionChatLogListReq) (*sessionm.SessionChatLogListResp, error) {
	if req.SessionId == "" {
		return nil, fmt.Errorf("sessionId is empty")
	}
	logs, _, err := s.dal.Db().Gdb().Urtyg_ai_agent().Ai_chat_logs().Gen().Find(&model.Ai_chat_logs_QueryReq{
		Content:   nil,
		CreatedAt: nil,
		ID:        nil,
		Order:     nil,
		Role:      nil,
		SessionID: &req.SessionId,
		UserID:    nil,
		OrderBy:   nil,
		IsLike:    false,
		Page:      nil,
	})
	if err != nil {
		return nil, err
	}
	// 排序逻辑
	sort.Slice(logs, func(i, j int) bool {
		// 按 Order 升序
		return logs[i].Order < logs[j].Order
	})
	for _, v := range logs {
		msg := &schema.Message{}
		err := json.Unmarshal([]byte(v.Content), msg)
		if err != nil {
			return nil, err
		}
		v.Content = msg.Content
	}
	res := &sessionm.SessionChatLogListResp{
		SessionId: req.SessionId,
		ChatLogs:  logs,
	}
	return res, nil
}

func (s *Service) DeleteSessions(ctx context.Context, req sessionm.DeleteSessionsReq) (*sessionm.DeleteSessionsResp, error) {
	if len(req.SessionIds) == 0 {
		return nil, fmt.Errorf("sessionIds is empty")
	}
	for _, v := range req.SessionIds {
		if v == "" {
			return nil, fmt.Errorf("sessionId is empty")
		}
	}
	result, err := s.dal.Db().Gdb().Urtyg_ai_agent().GenQ().Ai_chat_logs.
		WithContext(context.Background()).
		Where(s.dal.Db().Gdb().Urtyg_ai_agent().GenQ().Ai_chat_logs.SessionID.In(req.SessionIds...)).
		Delete()

	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return nil, nil
}

func (s *Service) DeleteChatLogs(ctx context.Context, req sessionm.DeleteChatLogsReq) (*sessionm.DeleteChatLogsResp, error) {
	dels := make([]*model.Ai_chat_logs, 0)
	for _, v := range req.IDs {
		dels = append(dels, &model.Ai_chat_logs{
			ID: v,
		})
	}
	err := s.dal.Db().Gdb().Urtyg_ai_agent().Ai_chat_logs().Gen().Del(dels...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Service) CreatSession(ctx context.Context, req sessionm.CreatSessionReq) (*sessionm.CreatSessionResp, error) {
	if req.UserId == "" {
		return nil, fmt.Errorf("userId is empty")
	}
	sessionId := fmt.Sprintf("%d", utils.GetIntID())
	session, err := s.uJwt.GetSession(req.UserId, sessionId)
	if err != nil {
		return nil, err
	}
	res := &sessionm.CreatSessionResp{
		SessionId: *session,
	}
	err = s.dal.Db().Gdb().Urtyg_ai_agent().Ai_session_logs().Gen().Save(&model.Ai_session_logs{
		ID:       res.SessionId,
		UserID:   req.UserId,
		PromptID: string(iconsts.PromptID_IntelligentAssistant),
		LastAt:   time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

var NoPromptTypeErr = errors.New("该会话类型提示词参数未设置,请联系管理员")

// QuerySessionChatLogsByUser 根据用户和既有的会话信息,返回会话的id和聊天记录,ut-ai-font 特殊应用接口
func (s *Service) QuerySessionChatLogsByUser(ctx context.Context, req sessionm.QuerySessionChatLogsByUserReq) (*sessionm.QuerySessionChatLogsByUserResp, error) {
	if req.UserId == "" {
		return nil, fmt.Errorf("userId is empty")
	}
	if req.PromptType == nil || *req.PromptType == "" {
		return nil, fmt.Errorf("promptType is empty")
	}

	res := &sessionm.QuerySessionChatLogsByUserResp{
		SessionId:  "",
		PromptType: "",
		ChatLogs:   make([]*sessionm.ChatLog, 0),
	}
	// 先查用户所有的会话记录
	// 查提示词模板id,根据type
	first, err := s.dal.Db().Gdb().Urtyg_ai_agent().Ai_prompt().Gen().First(&model.Ai_prompt{
		Type: *req.PromptType,
	})
	if err != nil {
		return nil, err
	}
	if first == nil {
		return nil, NoPromptTypeErr
	}
	orderBy1 := "last_at desc"
	sessionLogs, _, err := s.dal.Db().Gdb().Urtyg_ai_agent().Ai_session_logs().Gen().Find(&model.Ai_session_logs_QueryReq{
		ID:       nil,
		LastAt:   nil,
		PromptID: &first.ID,
		UserID:   &req.UserId,
		OrderBy:  &orderBy1,
		IsLike:   false,
		Page:     nil,
	})
	if err != nil {
		return nil, err
	}
	if len(sessionLogs) == 0 {
		// 新建一个会话id
		res.PromptType = *req.PromptType
		sessionId := fmt.Sprintf("%d", utils.GetIntID())
		session, err := s.uJwt.GetSession(req.UserId, sessionId)
		if err != nil {
			return nil, err
		}
		res.SessionId = *session
		err = s.dal.Db().Gdb().Urtyg_ai_agent().Ai_session_logs().Gen().Save(&model.Ai_session_logs{
			ID:       res.SessionId,
			UserID:   req.UserId,
			PromptID: first.ID,
			LastAt:   time.Now(),
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	} else {
		// 拿最后一个会话记录
		lastSession := sessionLogs[0]
		orderBy := "created_at asc"
		chatLogs, _, err := s.dal.Db().Gdb().Urtyg_ai_agent().Ai_chat_logs().Gen().Find(&model.Ai_chat_logs_QueryReq{
			Content:   nil,
			CreatedAt: nil,
			ID:        nil,
			Order:     nil,
			Role:      nil,
			SessionID: &lastSession.ID,
			UserID:    &req.UserId,
			OrderBy:   &orderBy,
			IsLike:    false,
			Page:      nil,
		})
		if err != nil {
			return nil, err
		}
		res.SessionId = lastSession.ID
		res.PromptType = first.Type
		if chatLogs != nil {
			data := make([]*sessionm.ChatLog, 0)
			for _, v := range chatLogs {
				msg := &schema.Message{}
				err := json.Unmarshal([]byte(v.Content), msg)
				if err != nil {
					return nil, err
				}
				v.Content = msg.Content
				v1 := sessionm.ChatLog{
					Role:      v.Role,
					Content:   v.Content,
					Timestamp: v.CreatedAt.Unix(),
				}
				data = append(data, &v1)
			}
			res.ChatLogs = data
		}
		return res, nil
	}
}

var _ If = &Service{}
