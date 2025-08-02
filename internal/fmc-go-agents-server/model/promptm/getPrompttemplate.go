package promptm

type GetPromptTemplateReq struct {
}
type GetPromptTemplateResp struct {
	List []GetPromptTemplateRespData `json:"list"`
}

type GetPromptTemplateRespData struct {
	Name        string `json:"name"`        // 提示词模板名称
	Description string `json:"description"` // 提示词模板描述
	Content     string `json:"content"`     // 提示词模板内容
}
