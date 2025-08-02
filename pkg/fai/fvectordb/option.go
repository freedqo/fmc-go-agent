package fvectordb

func NewDefaultOption() *Option {
	return &Option{
		Es8: Es8Option{
			Addresses: []string{"http://192.168.53.217:9200"},
			Username:  "elastic",
			Password:  "Unitech@1998",
		},
		IndexName: "utgdrag",
		Rewrite: &TOpenAiModel{
			APIKey:  "",
			BaseURL: "http://192.168.53.217:11434/v1",
			Model:   "qwen3:4b",
		},
		Qa: TOpenAiModel{
			APIKey:  "",
			BaseURL: "http://192.168.53.217:11434/v1",
			Model:   "qwen3:4b",
		},
		Embedding: TOpenAiModel{
			APIKey:  "",
			BaseURL: "http://192.168.53.217:11434/v1",
			Model:   "bge-m3:latest",
		},
	}
}

type Option struct {
	Es8       Es8Option     // es8客户端配置
	IndexName string        // 索引名称
	Embedding TOpenAiModel  // 嵌入向量模型
	Qa        TOpenAiModel  // QA对生成模型
	Rewrite   *TOpenAiModel // 重写模型
}

type Es8Option struct {
	Addresses []string // A list of Elasticsearch nodes to use.
	Username  string   // Username for HTTP Basic Authentication.
	Password  string   // Password for HTTP Basic Authentication.
}
type TOpenAiModel struct {
	APIKey  string `comment:"APIkey"`
	BaseURL string `comment:"BaseURL"`
	Model   string `comment:"Model"`
}
