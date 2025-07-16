package promptm

type GetPromptTemplateReq struct {
}
type GetPromptTemplateResp struct {
	List []GetPromptTemplateRespData `json:"list"`
}

type GetPromptTemplateRespData struct {
	Name        string `json:"name"`        // 模板名称
	Description string `json:"description"` // 模板描述
	Content     string `json:"content"`     // 模板内容
}
