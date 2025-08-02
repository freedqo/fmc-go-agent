package sessionm

type CreatSessionReq struct {
	UserId string `json:"userId"`
}
type CreatSessionResp struct {
	SessionId string `json:"sessionId"`
}
