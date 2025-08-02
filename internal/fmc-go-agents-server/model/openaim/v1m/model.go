package v1m

type ResModel struct {
	Id      string `json:"id"`
	Created int    `json:"created"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}
