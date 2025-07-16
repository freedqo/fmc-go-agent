package uaivectordb

func NewOption() *Option {
	return &Option{
		IRVModel: &IRVModelOption{
			BaseURL:      "http://192.168.53.217:11434",
			APIKey:       "ollama key",
			Model:        "bge-m3:latest",
			Organization: "ut-pc2-gd",
			Provider:     "ollama",
			Timeout:      120,
		},
		RedisStack: &RedisStack{
			Addr:      "192.168.53.217:16379",
			Protocol:  2,
			Dimension: 4096,
		},
		LoadMdFilePloy: &LoadMdFilePloy{
			IsLoadMdFiles: false,
			Dir:           "./knowdb/md",
		},
	}
}

type Option struct {
	IRVModel       *IRVModelOption `comment:"向量数据库使用的大模型配置"`
	RedisStack     *RedisStack     `comment:"向量数据库RedisStack配置"`
	LoadMdFilePloy *LoadMdFilePloy `comment:"本地知识文档（*.md）加载策略"`
}

type RedisStack struct {
	Addr      string `comment:"RedisStack地址,192.168.53.217:16789"` // 192.168.53.217:16789
	Protocol  int    `comment:"协议类型,2"`                            // 2
	Dimension int    `comment:"对齐格式,4096"`                         // 4096
	Db        int    `comment:"数据库索引,0"`                           // 0 int
}
type LoadMdFilePloy struct {
	IsLoadMdFiles bool   `comment:"是否加载本地文件到向量数据库"`
	Dir           string `comment:"本地md文件路径"`
}

type IRVModelOption struct {
	APIKey       string `comment:"API-秘钥"`    // API秘钥
	BaseURL      string `comment:"API-链接"`    // API基础链接
	Organization string `comment:"API-使用组织"`  // API使用组织
	Provider     string `comment:"API-模型提供商"` // API提供商-OpenAI、Ollama、DeepSeek等支持OpenAi Api的厂商
	Model        string `comment:"API-应用模型"`  // API-应用模型
	Timeout      int64  `comment:"API-超时时间"`  // API超时时间,单位秒
}
