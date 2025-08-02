package knowdbm

type GetFileListReq struct {
	Type     string `json:"type"`
	Page     int32  `json:"page"`
	PageSize int32  `json:"pageSize"`
}

type GetFileListResp struct {
	List []*TypeList `json:"data"`
}
type TypeList struct {
	Type     string       `json:"type"`
	FileList []*TFileInfo `json:"fileList"`
}
type TFileInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
	Size string `json:"size"`
	Date string `json:"date"`
}
