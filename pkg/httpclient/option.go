package httpclient

// Option 定义客户端配置
type Option struct {
	Enable     bool     `comment:"是否启用"`        // 是否启用
	BaseURL    []string `comment:"基础URL"`       // 基础URL
	Timeout    int      `comment:"请求超时时间,单位s"`  // 请求超时时间
	RetryCount int      `comment:"重试次数"`        // 重试次数
	RetryDelay int      `comment:"重试间隔时间,单位ms"` // 重试间隔时间
}

func NewDefaultOption() *Option {
	return &Option{
		BaseURL:    []string{"http://localhost:80"},
		Timeout:    60,
		RetryCount: 3,
		RetryDelay: 100,
	}
}
