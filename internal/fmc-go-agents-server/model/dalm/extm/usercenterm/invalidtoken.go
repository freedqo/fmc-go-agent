package usercenterm

type InvalidTokenReq struct {
	TokenId string `json:"token_id"`
}

type InvalidTokenResp struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	Total     int         `json:"total"`
	Timestamp int64       `json:"timestamp"`
	Success   bool        `json:"success"`
}
