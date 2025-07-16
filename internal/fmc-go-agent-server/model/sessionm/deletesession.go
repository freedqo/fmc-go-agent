package sessionm

type DeleteSessionsReq struct {
	SessionIds []string `json:"sessionIds"`
}

type DeleteSessionsResp struct {
}
