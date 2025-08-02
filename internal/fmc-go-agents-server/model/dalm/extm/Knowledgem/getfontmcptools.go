package Knowledgem

import "github.com/mark3labs/mcp-go/mcp"

type GetFontMCPToolsReq struct {
}

type GetFontMCPToolsResp struct {
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	Data []McpToolInfo `json:"data"`
}
type McpToolInfo struct {
	Name   string     // 工具名称
	Des    string     // 工具描述
	Object *ArgObject // 参数对象
	/*
		Title字段是一个字符串类型，该字段是可选的（即可以为空）。
		ReadOnlyHint字段是一个指向布尔类型的指针,如果该指针指向的布尔值为true，表示该工具不会修改其运行环境。
		DestructiveHint字段是一个指向布尔类型的指针，如果该指针指向的布尔值为true，意味着该工具可能会执行破坏性的更新操作。
		IdempotentHint字段是一个指向布尔类型的指针，如果该指针指向的布尔值为true，表明重复使用相同参数调用该工具不会产生额外的效果。
		OpenWorldHint字段是一个指向布尔类型的指针，如果该指针指向的布尔值为true，说明该工具会与外部实体进行交互。
	*/
	Annotations mcp.ToolAnnotation // 描述工具行为的可选属性
}

type ArgObject struct {
	IsRequired bool         // 是否必填
	Des        string       // 参数描述
	Properties []Properties // 参数列表
}
type Properties struct {
	Name      string      // 参数名称
	Type      string      //string, number, boolean, array
	ArrayItem *Properties // 数组类型参数的子项
	Des       string      // 参数描述
}
