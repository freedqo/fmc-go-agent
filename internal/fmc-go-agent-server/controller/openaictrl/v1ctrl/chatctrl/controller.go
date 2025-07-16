package chatctrl

import (
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/schema"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/openaim/v1m/chatm"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"
)

func New(service service.If) If {
	return &Controller{
		service: service,
	}
}

type Controller struct {
	service service.If
}

// Completions 与Ai Agent聊天
//
//	@Summary		与Ai Agent聊天
//	@Description	与Ai Agent聊天
//	@Tags			Openai Api 接口管理
//	@Accept			json
//	@Produce		json
//	@Param			Tokenid						header		string						true	"Tokenid 用户登录令牌"
//	@Param			SessionId					header		string						true	"用户与模型对话令牌"
//	@Param			chatm.ChatCompletionsReq	body		chatm.ChatCompletionsReq	true	"请求参数"
//	@Success		200							{object}	webapp.Response{}			"成功响应"
//	@Router			/openai/v1/chat/completions [post]
func (ctrl *Controller) Completions(c *gin.Context) {
	// 创建一个新的ChatCompletion请求参数
	req := chatm.ChatCompletionsReq{}
	// 绑定请求参数到body
	err := c.ShouldBind(&req)
	// 如果绑定失败，返回400错误
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// 从请求参数中获取sessionId和userId，如果不存在则生成一个新的UUID
	sessionId := c.GetHeader("SessionId")
	if sessionId == "" {
		c.JSON(400, gin.H{"error": "sessionId is required"})
		return
	}
	//
	tokenId := c.GetHeader("Tokenid")
	if tokenId == "" {
		c.JSON(400, gin.H{"error": "Tokenid is required"})
		return
	}
	c.Writer.Header()

	// 调用OpenAI的ChatCompletion接口，传入请求参数和上下文
	res, err := ctrl.service.OpenAi().V1().Chat().Completions(c.Request.Context(), sessionId, req)
	// 如果调用失败，返回500错误
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	id, _ := uuid.NewUUID()
	var index int64

	switch res.(type) {
	case *schema.StreamReader[*schema.Message]:
		// 设置响应头
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

		// 强制刷新头信息
		c.Writer.(http.Flusher).Flush()
		msgs := res.(*schema.StreamReader[*schema.Message])
		for {
			v, err := msgs.Recv()
			index++
			if err != nil {
				if err == io.EOF {
					// 流结束，发送空行表示完成
					c.Writer.Write([]byte("data: [DONE]\n\n"))
				} else {
					// 错误处理
					c.Writer.Write([]byte(fmt.Sprintf("data: {\"error\": \"%s\"}\n\n", err.Error())))
				}
				c.Writer.(http.Flusher).Flush()
				break
			}
			// 提取delta内容并格式化为SSE
			if len(v.Content) > 0 {
				resMsg := chatm.ResChatCompletionsStream{
					Choices:           make([]chatm.Choices, 0),
					Created:           int(time.Now().Unix()),
					ID:                id.String(),
					Object:            "chat.completion.chunk",
					Model:             req.Model,
					SystemFingerprint: "fp_8802369eaa_prod0623_fp8_kvcache",
					Usage:             nil,
				}
				if v.ResponseMeta != nil && v.ResponseMeta.Usage != nil {
					resMsg.Usage = &chatm.Usage{
						PromptTokens:          v.ResponseMeta.Usage.PromptTokens,
						CompletionTokens:      v.ResponseMeta.Usage.CompletionTokens,
						TotalTokens:           v.ResponseMeta.Usage.TotalTokens,
						PromptTokensDetails:   nil,
						PromptCacheHitTokens:  0,
						PromptCacheMissTokens: 0,
					}
				}
				choice := chatm.Choices{
					Index: int(index),
					Delta: chatm.Delta{
						Content: v.Content,
					},
				}
				if v.ResponseMeta != nil {
					choice.Logprobs = v.ResponseMeta.LogProbs
					choice.FinishReason = v.ResponseMeta.FinishReason
				}
				resMsg.Choices = append(resMsg.Choices, choice)
				str, err := json.Marshal(resMsg)
				if err != nil {
					return
				}
				c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", str)))

				c.Writer.(http.Flusher).Flush()

				// 检查客户端连接是否关闭
				if c.Request.Context().Err() != nil {
					break
				}
				log.SysLog().Infof("out stream index: %d,content: %s", index, v.Content)
			}
		}
		break
	case *schema.Message:
		{
			v1, ok := res.(*schema.Message)
			if !ok {
				c.JSON(500, gin.H{"error": "unknown type"})
			}
			resMsg := chatm.ResChatCompletions{
				Choices: make([]chatm.Choice, 0),
				Created: time.Now().Unix(),
				ID:      id.String(),
				Object:  "chat.completion",
				Usage:   nil,
			}
			if v1.ResponseMeta != nil && v1.ResponseMeta.Usage != nil {
				resMsg.Usage = &chatm.ChatCompletionsUsage{
					CompletionTokens: v1.ResponseMeta.Usage.CompletionTokens,
					PromptTokens:     v1.ResponseMeta.Usage.PromptTokens,
					TotalTokens:      v1.ResponseMeta.Usage.TotalTokens,
				}
			}

			resMsg.Choices = append(resMsg.Choices, chatm.Choice{
				FinishReason: &v1.ResponseMeta.FinishReason,
				Index:        &index,
				Message: &chatm.Message{
					Content: v1.Content,
					Role:    string(v1.Role),
				},
			})
			c.JSON(200, resMsg)
			break
		}
	default:
		{
			c.JSON(500, gin.H{"error": "unknown type"})
			break
		}
	}
	return
}

func (ctrl *Controller) Models(c *gin.Context) {
	models, err := ctrl.service.OpenAi().V1().Models(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, models)
}

var _ If = &Controller{}
