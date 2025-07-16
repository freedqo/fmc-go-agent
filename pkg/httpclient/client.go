package httpclient

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// NewHttpClient 创建一个新的HTTP客户端实例
func NewHttpClient(serverName string, config *Option) *HttpClient {
	if config == nil {
		config = NewDefaultOption()
	}
	if config.BaseURL == nil || len(config.BaseURL) == 0 {
		panic("http客户端,请求的URL不能为空")
	}
	if serverName == "" {
		panic("http客户端,请求的服务器名称不能为空")
	}
	if config.Timeout <= 0 {
		config.Timeout = 60
	}
	if config.RetryCount <= 0 {
		config.RetryCount = 1
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = 100
	}

	return &HttpClient{
		ServerName: serverName,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		mu:              sync.RWMutex{},
		Config:          config,
		middlewareChain: nil,
	}
}

// HttpClient 定义HTTP客户端结构
type HttpClient struct {
	ServerName      string
	mu              sync.RWMutex
	client          *http.Client
	Config          *Option
	middlewareChain MiddlewareFunc
	baseUrlIndex    int
}

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// HandlerFunc 处理器函数类型
type HandlerFunc func(ctx context.Context, req *http.Request) (*http.Response, error)

// Use 添加中间件到中间件链
func (c *HttpClient) Use(middleware MiddlewareFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.middlewareChain == nil {
		c.middlewareChain = middleware
		return
	}
	oldMiddleware := c.middlewareChain
	c.middlewareChain = func(next HandlerFunc) HandlerFunc {
		return middleware(oldMiddleware(next))
	}
}
