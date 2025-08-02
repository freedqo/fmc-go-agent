package v1m

type ReqEmbeddings struct {
	// 输入文本以获取嵌入，编码为字符串或标记数组。要在单个请求中获取多个输入的嵌入，请传递一个字符串数组或令牌数组数组。每个输入的长度不得超过 8192 个标记。
	Input string `json:"input"`
	// 要使用的模型的 ID。您可以使用[List models](https://platform.openai.com/docs/api-reference/models/list)
	// API 来查看所有可用模型，或查看我们的[模型概述](https://platform.openai.com/docs/models/overview)以了解它们的描述。
	Model string `json:"model"`
}
type ResEmbeddings struct {
	Data   []Datum         `json:"data"`
	Model  string          `json:"model"`
	Object string          `json:"object"`
	Usage  EmbeddingsUsage `json:"usage"`
}

type Datum struct {
	Embedding []float64 `json:"embedding,omitempty"`
	Index     *int64    `json:"index,omitempty"`
	Object    *string   `json:"object,omitempty"`
}

type EmbeddingsUsage struct {
	PromptTokens int64 `json:"prompt_tokens"`
	TotalTokens  int64 `json:"total_tokens"`
}
