package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

// Request 请求参数结构体
type Request struct {
	Headers map[string]string
	Query   map[string]interface{}
	Body    interface{}
}

// Response 响应参数结构体
type Response struct {
	Headers    map[string][]string
	StatusCode int
	Status     string
	Body       interface{}
}

// NewRequest 创建请求参数
func (c *HttpClient) NewRequest(reqBody interface{}) Request {
	return Request{
		Headers: make(map[string]string),
		Query:   make(map[string]interface{}),
		Body:    reqBody,
	}
}

// NewResponse 创建响应参数
func (c *HttpClient) NewResponse(resBody interface{}) *Response {
	return &Response{
		Headers:    make(map[string][]string),
		StatusCode: 500,
		Status:     "请求服务异常",
		Body:       resBody,
	}
}

// Get 发送GET请求
// 入参： ctx 请求上下文
// 入参： path 请求路径
// 入参： request 请求参数
// 入参： response 响应参数
// 返回： error 错误信息
func (c *HttpClient) Get(ctx context.Context, path string, request Request, response *Response) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	err := c.do(ctx, http.MethodGet, path, request, response)
	if err != nil {
		return err
	}

	return nil
}

// Post 发送POST请求
// 入参： ctx 请求上下文
// 入参： path 请求路径
// 入参： request 请求参数
// 入参： response 响应参数
// 返回： error 错误信息
func (c *HttpClient) Post(ctx context.Context, path string, request Request, response *Response) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	err := c.do(ctx, http.MethodPost, path, request, response)
	if err != nil {
		return err
	}

	return nil
}

// Put 发送PUT请求
// 入参： ctx 请求上下文
// 入参： path 请求路径
// 入参： request 请求参数
// 入参： response 响应参数
// 返回： error 错误信息
func (c *HttpClient) Put(ctx context.Context, path string, request Request, response *Response) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	err := c.do(ctx, http.MethodPut, path, request, response)
	if err != nil {
		return err
	}

	return nil
}

// Delete 发送DELETE请求
// 入参： ctx 请求上下文
// 入参： path 请求路径
// 入参： request 请求参数
// 入参： response 响应参数
// 返回： error 错误信息
func (c *HttpClient) Delete(ctx context.Context, path string, request Request, response *Response) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	err := c.do(ctx, http.MethodDelete, path, request, response)
	if err != nil {
		return err
	}

	return nil
}

// setHeaders 将 Request 结构体的 Headers 参数设置到 http.Request 的请求头中
// 入参： req http.Request 请求对象
// 入参： headers map[string]string 请求头参数
// 返回： error 错误信息
func (c *HttpClient) setHeaders(req *http.Request, headers map[string]string) {
	if headers == nil {
		headers = make(map[string]string)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

// setQuery 将 Request 结构体的 Query 参数转换为标准的查询参数
// 入参： req http.Request 请求对象
// 入参： query map[string]interface{} 查询参数
// 返回： error 错误信息
func (c *HttpClient) setQuery(req *http.Request, query map[string]interface{}) {
	q := req.URL.Query()
	for k, v := range query {
		switch value := v.(type) {
		case string:
			q.Add(k, value)
		case int:
			q.Add(k, fmt.Sprintf("%d", value))
		case bool:
			q.Add(k, fmt.Sprintf("%v", value))
		case []string:
			for _, s := range value {
				q.Add(k, s)
			}
		default:
			// 其他类型转换为字符串
			q.Add(k, fmt.Sprintf("%v", value))
		}
	}
	req.URL.RawQuery = q.Encode()
}

// NewMultipartWriter 创建一个multipart.Writer
// 入参： 无
// 返回： *MultipartWriter multipart.Writer 指针
func (c *HttpClient) NewMultipartWriter() *MultipartWriter {
	mw := MultipartWriter{
		Buf:    bytes.Buffer{},
		Writer: nil,
	}
	mw.Writer = multipart.NewWriter(&mw.Buf)
	return &mw
}

type MultipartWriter struct {
	Buf    bytes.Buffer
	Writer *multipart.Writer
}

// setBody 将 Request 结构体的 Body 参数设置到 http.Request 的请求体中
// 入参： req http.Request 请求对象
// 入参： body interface{} 请求体参数
// 返回： error 错误信息
func (c *HttpClient) setBody(req *http.Request, body interface{}) error {
	var bodyBytes []byte
	var err error
	var contentType string

	switch v := body.(type) {
	case []byte:
		bodyBytes = v
		contentType = "application/octet-stream"
	case string:
		bodyBytes = []byte(v)
		contentType = "text/plain"
	case url.Values:
		bodyBytes = []byte(v.Encode())
		contentType = "application/x-www-form-urlencoded"
	case map[string]interface{}:
		bodyBytes, err = json.Marshal(v)
		if err != nil {
			return errors.New("请求体序列化失败")
		}
		contentType = "application/json"
	case *MultipartWriter:
		// 关闭multipart.Writer以完成数据写入
		err = v.Writer.Close()
		if err != nil {
			return fmt.Errorf("关闭multipart.Writer失败: %w", err)
		}
		bodyBytes = v.Buf.Bytes()
		contentType = v.Writer.FormDataContentType()
	default:
		bodyBytes, err = json.Marshal(v)
		if err != nil {
			return errors.New("请求体序列化失败")
		}
		contentType = "application/json"
	}
	req.Header.Set("Content-Type", contentType)
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	req.ContentLength = int64(len(bodyBytes))
	return nil
}

// do 发送HTTP请求并返回响应体
// 入参： ctx 请求上下文
// 入参： method string 请求方法
// 入参： path string 请求路径
// 入参： request *Request 请求参数
// 入参： response *Response 响应对象指针
// 返回： error 错误信息
func (c *HttpClient) do(ctx context.Context, method string, path string, request Request, response *Response) error {
	var resp *http.Response
	var err error

	// 设置重试机制
	for i := 0; i <= c.Config.RetryCount; i++ {
		if i > 0 {
			time.Sleep(time.Duration(c.Config.RetryDelay) * time.Millisecond)
		}
		if i > 1 {
			c.baseUrlIndex += 1
			if c.baseUrlIndex >= len(c.Config.BaseURL) {
				c.baseUrlIndex = 0
			}
		}
		nowUrl := fmt.Sprintf("%s%s", c.Config.BaseURL[c.baseUrlIndex], path)

		req, err := http.NewRequestWithContext(ctx, method, nowUrl, nil)
		if err != nil {
			return err
		}
		// 设置默认请求头
		c.setHeaders(req, request.Headers)

		// 设置查询参数
		c.setQuery(req, request.Query)

		// 设置请求体
		err = c.setBody(req, request.Body)
		if err != nil {
			return err
		}

		// 如果没有中间件链，则直接发送请求
		if c.middlewareChain == nil {
			resp, err = c.client.Do(req)
		} else {
			handler := c.middlewareChain(func(ctx context.Context, r *http.Request) (*http.Response, error) {
				return c.client.Do(r)
			})
			resp, err = handler(ctx, req)
		}

		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("响应为空")
	}
	if response == nil {
		return errors.New("响应接收对象为空")
	}
	response.Headers = resp.Header
	response.StatusCode = resp.StatusCode
	response.Status = resp.Status

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 根据 response.Body 指针指向的类型来判断是接收结构体还是文件
	switch target := response.Body.(type) {
	case *[]byte:
		// 期望接收文件，直接将字节切片赋值给指针指向的变量
		*target = body
	default:
		// 尝试进行 JSON 反序列化
		if err := json.Unmarshal(body, target); err != nil {
			return errors.New(fmt.Sprintf("JSON 反序列化错误: %s,原始Body:%s", err.Error(), string(body)))
		}
	}
	return nil
}

// do 发送HTTP请求并返回响应体
// 入参： ctx 请求上下文
// 入参： req http.Request 请求对象
// 入参： response *Response 响应对象指针
// 返回： error 错误信息
func (c *HttpClient) do1(ctx context.Context, req *http.Request, response *Response) error {
	var resp *http.Response
	var err error
	// 设置重试机制
	for i := 0; i <= c.Config.RetryCount; i++ {
		if i > 0 {
			time.Sleep(time.Duration(c.Config.RetryDelay) * time.Millisecond)
		}

		// 如果没有中间件链，则直接发送请求
		if c.middlewareChain == nil {
			resp, err = c.client.Do(req)
		} else {
			handler := c.middlewareChain(func(ctx context.Context, r *http.Request) (*http.Response, error) {
				return c.client.Do(r)
			})
			resp, err = handler(ctx, req)
		}

		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}
	response.Headers = resp.Header
	response.StatusCode = resp.StatusCode
	response.Status = resp.Status

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 根据 response.Body 指针指向的类型来判断是接收结构体还是文件
	switch target := response.Body.(type) {
	case *[]byte:
		// 期望接收文件，直接将字节切片赋值给指针指向的变量
		*target = body
	default:
		// 尝试进行 JSON 反序列化
		if err := json.Unmarshal(body, target); err != nil {
			return errors.New(fmt.Sprintf("JSON 反序列化错误: %s,原始Body:%s", err.Error(), string(body)))
		}
	}
	return nil
}
