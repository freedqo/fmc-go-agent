package faicharmodel

const (
	OpenAI   ModelProvider = "OpenAI"
	Ollama   ModelProvider = "Ollama"
	DeepSeek ModelProvider = "DeepSeek"
	Qwen     ModelProvider = "Qwen"
)

var SupportModelProvider = map[ModelProvider]string{
	OpenAI:   "OpenAI",
	Ollama:   "Ollama",
	DeepSeek: "DeepSeek",
	Qwen:     "Qwen",
}

type Option struct {
	APIKey       string `comment:"API-秘钥"`                                                // API秘钥
	BaseURL      string `comment:"API-链接"`                                                // API基础链接
	Organization string `comment:"API-使用组织"`                                              // API使用组织
	Provider     string `comment:"API-模型提供商:OpenAI、Ollama、DeepSeek,Qwen等支持OpenAi Api的厂商"` // API提供商:OpenAI、Ollama、DeepSeek等支持OpenAi Api的厂商
	Model        string `comment:"API-应用模型"`                                              // API-应用模型
	Timeout      int64  `comment:"API-超时时间"`                                              // API超时时间,单位秒
}
