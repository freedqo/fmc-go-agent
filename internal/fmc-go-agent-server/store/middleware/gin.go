package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/log"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/ujwtx"
	"github.com/freedqo/fmc-go-agent/pkg/utils"
	"github.com/freedqo/fmc-go-agent/pkg/webapp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
	"time"
)

// GinJWTAuthMiddleware JWT 认证中间件
// 入参：无
// 出参：gin.HandlerFunc 中间件函数
func GinJWTAuthMiddleware() gin.HandlerFunc {
	jwtWrapper := ujwtx.GetJWT()
	return func(c *gin.Context) {
		token := c.Request.Header.Get("tokenId")
		if token == "" {
			c.JSON(http.StatusUnauthorized, webapp.Response{
				Code:    http.StatusUnauthorized,
				Message: "请求头未提供tokenId",
				Data:    nil,
			})
			c.Abort()
			return
		}
		// 处理没有 Bearer 前缀的情况
		var tokenStr string
		parts := strings.Split(token, " ")
		if len(parts) == 1 {
			tokenStr = parts[0]
		} else if len(parts) == 2 && parts[0] == "Bearer" {
			tokenStr = parts[1]
		} else {
			c.JSON(http.StatusUnauthorized, webapp.Response{
				Code:    http.StatusUnauthorized,
				Message: "请求头tokenId格式错误",
				Data:    nil,
			})
			c.Abort()
			return
		}
		claims, err := jwtWrapper.VerifyToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, webapp.Response{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
				Data:    nil,
			})
			c.Abort()
			return
		}
		// 将用户信息存储到上下文，方便后续路由处理函数使用
		c.Set("userName", claims.UserName)
		c.Set("userId", claims.UserId)
		// 继续处理后续的路由逻辑
		c.Next()
	}
}

// GinRecoveryMiddleware 中间件
// 入参：无
// 出参：gin.HandlerFunc 中间件函数
func GinRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 生成一个唯一的错误ID，用于后续的错误跟踪
				PanicId := uuid.New().ID()
				// 记录 panic 信息
				log.SysLog().Errorf("PanicId: %d, Panic: %v", PanicId, err)
				// 打印堆栈跟踪信息（可选）
				log.SysLog().Errorf(utils.StackSkip(1, -1))

				// 返回错误响应
				c.JSON(http.StatusInternalServerError, gin.H{
					"id":      500,
					"error":   "服务异常,请联系管理员," + fmt.Sprintf("PanicId: %d, Panic: %v", PanicId, err),
					"result":  nil,
					"code":    500,
					"message": "服务异常,请联系管理员," + fmt.Sprintf("PanicId: %d, Panic: %v", PanicId, err),
					"data":    nil,
				})
			}
		}()
		c.Next()
	}
}

// GinLoggerMiddleware 中间件：记录请求日志
// 入参：无
// 出参：gin.HandlerFunc 中间件函数
func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqInfo := ""
		start := time.Now()
		// 创建一个新的响应记录器
		responseWriter := &ResponseWriterWrapper{body: []byte{}, ResponseWriter: c.Writer}
		// 将新的响应记录器设置为当前上下文的 Writer
		c.Writer = responseWriter

		isKnife4jgo := strings.Contains(c.Request.URL.Path, "knife4jgo")
		if !isKnife4jgo {
			// 记录请求信息
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			var inBody interface{}
			_ = json.Unmarshal(bodyBytes, &inBody)
			bodyStr, _ := json.Marshal(inBody)

			reqInfo = fmt.Sprintf("请求内容：请求头: %+v, 查询: %+v, Body: %s", c.Request.Header, c.Request.URL.Query(), string(bodyStr))
		}

		// 执行请求
		c.Next()
		// 记录响应信息
		responseBody := responseWriter.body
		if !isKnife4jgo {
			resInfo := fmt.Sprintf("响应内容: 状态：%d, ", c.Writer.Status())
			msg := fmt.Sprintf("%s-->服务 %s %s [ %s ] %s;%s", c.Request.RemoteAddr, c.Request.Method, c.Request.URL.Path, time.Since(start).String(), reqInfo, resInfo)
			body := string(responseBody)
			if len(body) > 548576 {
				msg += "Body[to long...]: " + body[:1000] + " ... ..."
			} else {
				msg += "Body: " + body
			}
			if time.Since(start) > 800*time.Millisecond || c.Writer.Status() != 200 {
				log.SysLog().Warnf(msg)
			} else {
				log.SysLog().Infof(msg)
			}
		}
	}
}

// ResponseWriterWrapper 自定义一个包装器实现 gin.ResponseWriter 并重写 Write 方法
type ResponseWriterWrapper struct {
	body []byte
	gin.ResponseWriter
}

// 重写 Write 方法，保存响应体
// 入参：data []byte 响应数据
// 出参：int 写入的字节数
// 出参：error 错误信息
func (w *ResponseWriterWrapper) Write(data []byte) (int, error) {
	w.body = data
	gin.Recovery()
	return w.ResponseWriter.Write(data)
}
