package dbconversation

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/schema"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal/db/dbif"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/dalm/dbm/urtyg_ai_agent/model"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/log"
	"github.com/freedqo/fmc-go-agent/pkg/ujwt"
	"github.com/freedqo/fmc-go-agent/pkg/utils"
	"go.uber.org/zap"
	"sort"
	"time"
)

func New(sessionId string, uJwt ujwt.If, db dbif.If) If {
	uClaims, err := uJwt.VerifySession(sessionId)
	if err != nil {
		panic(err)
	}
	if uClaims == nil {
		panic("invalid session")
	}
	// 判断会话是否存在
	first, err := db.Urtyg_ai_agent().Ai_session_logs().Gen().First(&model.Ai_session_logs{
		ID: sessionId,
	})
	if err != nil {
		panic(err)
	}
	if first == nil {
		panic("会话Id无效,未授权会话,请重新创建会话")
	}
	if first.UserID != uClaims.UserId {
		panic("会话Id无效,用户ID与会话创建者不一致,请重新创建会话")
	}
	if first.PromptID == "" {
		panic("会话Id无效,未配置提示词,请重新创建会话")
	}
	first1, err1 := db.Urtyg_ai_agent().Ai_prompt().Gen().First(&model.Ai_prompt{
		ID: first.PromptID,
	})
	if err1 != nil {
		panic(err)
	}
	return &Service{
		sessionId: sessionId,
		userId:    uClaims.UserId,
		promptID:  first.PromptID,
		prompt:    first1.Content,
		uJwt:      uJwt,
		db:        db,
		log:       log.SysLog(),
	}
}

type Service struct {
	sessionId string
	userId    string
	promptID  string
	prompt    string
	db        dbif.If
	uJwt      ujwt.If
	log       *zap.SugaredLogger
}

func (s *Service) Append(msg *schema.Message) {
	js, err := json.Marshal(msg)
	if err != nil {
		s.log.Errorw("marshal message failed", "err", err)
		return
	}
	// 更新会话记录表的时间
	err = s.db.Urtyg_ai_agent().Ai_session_logs().Gen().Save(&model.Ai_session_logs{
		ID:       s.sessionId,
		UserID:   s.userId,
		PromptID: s.promptID,
		LastAt:   time.Now(),
	})
	if err != nil {
		s.log.Errorw("get max order failed", "err", err)
		return
	}
	order, err := s.db.Urtyg_ai_agent().Ai_chat_logs().Self().GetMaxOrder(context.Background(), s.sessionId)
	if err != nil {
		s.log.Errorw("get max order failed", "err", err)
		return
	}
	now := time.Now()
	err = s.db.Urtyg_ai_agent().Ai_chat_logs().Gen().Add(&model.Ai_chat_logs{
		ID:        utils.GetStringID(),
		SessionID: s.sessionId,
		UserID:    s.userId,
		Role:      string(msg.Role),
		Content:   string(js),
		Order:     int32(order + 1),
		CreatedAt: &now,
	})
	if err != nil {
		s.log.Errorw("append message failed", "err", err)
		return
	}
}

func (s *Service) GetMessages() []*schema.Message {
	chatLogs, _, err := s.db.Urtyg_ai_agent().Ai_chat_logs().Gen().Find(&model.Ai_chat_logs_QueryReq{
		Content:   nil,
		CreatedAt: nil,
		ID:        nil,
		Order:     nil,
		Role:      nil,
		SessionID: &s.sessionId,
		UserID:    &s.userId,
		OrderBy:   nil,
		IsLike:    false,
		Page:      nil,
	})
	if err != nil {
		return nil
	}
	sort.Slice(chatLogs, func(i, j int) bool {
		return chatLogs[i].Order < chatLogs[j].Order
	})
	res := make([]*schema.Message, 0, len(chatLogs))
	for _, v := range chatLogs {
		var msg schema.Message
		err := json.Unmarshal([]byte(v.Content), &msg)
		if err != nil {
			s.log.Errorw("unmarshal message failed", "err", err)
			continue
		}
		res = append(res, &msg)
	}
	return res
}

func (s *Service) Load() error {
	//TODO implement me
	panic("implement me")
}

func (s *Service) Save(msg *schema.Message) {
	//TODO implement me
	panic("implement me")
}

func (s *Service) GetPrompt() string {
	first, err := s.db.Urtyg_ai_agent().Ai_session_logs().Gen().First(&model.Ai_session_logs{
		ID: s.sessionId,
	})
	if err != nil {
		return ""
	}
	first1, err1 := s.db.Urtyg_ai_agent().Ai_prompt().Gen().First(&model.Ai_prompt{
		ID: first.PromptID,
	})
	if err1 != nil {
		return ""
	}
	return first1.Content
}
func (s *Service) GetSessionId() string {
	return s.sessionId
}

var _ If = &Service{}
