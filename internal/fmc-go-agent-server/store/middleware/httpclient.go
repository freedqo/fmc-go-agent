package middleware

import (
	"bytes"
	"context"
	"fmt"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/log"
	"github.com/freedqo/fmc-go-agent/pkg/httpclient"
	"io"
	"net/http"
	"time"
)

// HttpClientLoggerMiddleware 日志中间件
func HttpClientLoggerMiddleware(clientCfg *httpclient.HttpClient) httpclient.MiddlewareFunc {
	return func(next httpclient.HandlerFunc) httpclient.HandlerFunc {
		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
			start := time.Now()
			// 记录请求信息
			requestBody, _ := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(requestBody)) // 恢复请求体，以便后续使用

			reqInfo := fmt.Sprintf("请求头: %v ,请求参数: %s ,请求体: %s", req.Header, req.URL.RawQuery, string(requestBody))
			// 执行后续处理
			resp, err := next(ctx, req)
			if err != nil {
				return resp, err
			}
			// 记录响应信息
			responseBody, _ := io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewBuffer(responseBody)) // 恢复响应体，以便后续使用
			resInfo := fmt.Sprintf("响应头: %v ,", resp.Header)
			body := string(responseBody)
			if len(body) > 200000 {
				resInfo += "响应体: [to long...]: " + body[:1000] + " ... ..."
			} else {
				resInfo += "响应体: " + body
			}
			msg := fmt.Sprintf("本地请求-->[%s] %s %s [ %s ] %s ; %s ", clientCfg.ServerName, req.Method, req.URL.Scheme+"://"+req.URL.Host+req.URL.Path, time.Since(start), reqInfo, resInfo)
			if resp.StatusCode != 200 || time.Since(start) > 2*time.Second {
				log.SysLog().Warnf(msg)
			} else {
				log.SysLog().Infof(msg)
			}
			return resp, nil
		}
	}
}
