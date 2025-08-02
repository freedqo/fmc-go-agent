package sessionm

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/dbm"
	"time"
)

type UserSessionListReq struct {
	UserId string    `json:"userId"`
	Page   *dbm.Page `json:"page"`
}
type UserSessionListResp struct {
	SessionList []*UserSessionListRespData `json:"sessionList"`
	Page        *dbm.Page                  `json:"page"`
}
type UserSessionListRespData struct {
	SessionId string     `json:"sessionId"`
	Title     string     `json:"title"`
	CreatAt   *time.Time `json:"creatAt"`
}
