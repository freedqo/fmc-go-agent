package uaicharmodel

import (
	"context"
	"errors"
	edeepseek "github.com/cloudwego/eino-ext/components/model/deepseek"
	eollama "github.com/cloudwego/eino-ext/components/model/ollama"
	eopenai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"sync"
	"time"
)

func New(ctx context.Context, opt *Option) If {
	openaiClient := openai.NewClient(
		option.WithBaseURL(opt.BaseURL+"/v1"),     // 设置 OpenAI API 的基础 URL
		option.WithAPIKey(opt.APIKey),             // 设置 OpenAI API 的密钥
		option.WithOrganization(opt.Organization), // 设置 OpenAI API 的调用组织 ID
	)
	// 测试连接是否正常
	_, err := openaiClient.Models.List(ctx)
	if err != nil {
		panic(err)
	}

	// 创建一个 UChatModel 实例
	uCM := &UChatModel{
		mu:     sync.RWMutex{},
		ctx:    ctx,
		opt:    opt,
		openai: &openaiClient,
	}
	// 设置默认Eino的cm
	err = uCM.SetProvider(ModelProvider(opt.Provider))
	if err != nil {
		panic(err)
	}
	return uCM
}

type UChatModel struct {
	mu     sync.RWMutex    // 读写锁
	ctx    context.Context // 上下文
	opt    *Option         // 配置参数
	openai *openai.Client  // openai 客户端
	einoCm UChatModelIf    // eino 客户端
}

func (m *UChatModel) BindTools(tools []*schema.ToolInfo) error {
	return m.einoCm.BindTools(tools)
}

func (m *UChatModel) V1() *openai.Client {
	return m.openai
}

func (m *UChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if opts != nil {
		return m.einoCm.Generate(ctx, input, opts...)
	} else {
		return m.einoCm.Generate(ctx, input)
	}
}

func (m *UChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.einoCm.Stream(ctx, input, opts...)
}

func (m *UChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.einoCm.WithTools(tools)
}

// SetProvider 设置模型提供者,主要用于切换eino的cm
func (m *UChatModel) SetProvider(provider ModelProvider) error {
	// 检查提供的模型提供者是否在支持列表中
	_, ok := SupportModelProvider[provider]
	if !ok {
		// 如果不在支持列表中，返回错误
		return errors.New("unsupported model provider")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	// 根据提供的模型提供者，执行相应的操作
	switch provider {
	case OpenAI:
		{
			// 如果是OpenAI，执行相应的操作
			c, err := eopenai.NewChatModel(m.ctx, &eopenai.ChatModelConfig{
				APIKey:               m.opt.APIKey,
				Timeout:              time.Duration(m.opt.Timeout) * time.Second,
				HTTPClient:           nil,
				ByAzure:              false,
				AzureModelMapperFunc: nil,
				BaseURL:              m.opt.BaseURL,
				APIVersion:           "",
				Model:                m.opt.Model,
				MaxTokens:            nil,
				Temperature:          nil,
				TopP:                 nil,
				Stop:                 nil,
				PresencePenalty:      nil,
				ResponseFormat:       nil,
				Seed:                 nil,
				FrequencyPenalty:     nil,
				LogitBias:            nil,
				User:                 nil,
				ExtraFields:          nil,
			})
			if err != nil {
				return err
			}
			m.einoCm = c
			break
		}
	case Ollama:
		{
			// 如果是Ollama，执行相应的操作
			c, err := eollama.NewChatModel(m.ctx, &eollama.ChatModelConfig{
				BaseURL:    m.opt.BaseURL,
				Timeout:    time.Duration(m.opt.Timeout) * time.Second,
				HTTPClient: nil,
				Model:      m.opt.Model,
				Format:     nil,
				KeepAlive:  nil,
				Options:    nil,
			})
			if err != nil {
				return err
			}
			m.einoCm = c
			break
		}
	case DeepSeek:
		{
			// 如果是DeepSeek，执行相应的操作
			// 如果是Ollama，执行相应的操作
			c, err := edeepseek.NewChatModel(m.ctx, &edeepseek.ChatModelConfig{
				APIKey:             m.opt.APIKey,
				Timeout:            time.Duration(m.opt.Timeout) * time.Second,
				HTTPClient:         nil,
				BaseURL:            m.opt.BaseURL,
				Path:               "",
				Model:              m.opt.Model,
				MaxTokens:          0,
				Temperature:        0,
				TopP:               0,
				Stop:               nil,
				PresencePenalty:    0,
				ResponseFormatType: "",
				FrequencyPenalty:   0,
				LogProbs:           false,
				TopLogProbs:        0,
			})
			if err != nil {
				return err
			}
			m.einoCm = c
			break
		}
	default:
		{
			// 如果提供的模型提供者不在支持列表中，返回错误
			return errors.New("unsupported model provider")
		}
	}
	// 返回nil，表示设置成功
	return nil
}

var _ If = &UChatModel{}
