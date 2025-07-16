package mem

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/schema"
	"os"
	"sync"
)

type Conversation struct {
	mu            sync.Mutex
	ID            string            `json:"id"`
	Messages      []*schema.Message `json:"messages"`
	filePath      string
	maxWindowSize int
}

func (c *Conversation) GetPrompt() string {
	//TODO implement me
	panic("implement me")
}

func (c *Conversation) GetSessionId() string {
	//TODO implement me
	panic("implement me")
}

var _ ConversationIf = &Conversation{}

// Append 方法用于向 Conversation 结构体中添加一条消息
func (c *Conversation) Append(msg *schema.Message) {
	// 加锁，防止并发访问
	c.mu.Lock()
	// 在函数结束时解锁
	defer c.mu.Unlock()

	// 将消息添加到 Messages 切片中
	c.Messages = append(c.Messages, msg)

	// 保存消息
	c.Save(msg)
}

// GetFullMessages 获取完整的消息
func (c *Conversation) GetFullMessages() []*schema.Message {
	// 加锁
	c.mu.Lock()
	// 在函数结束时解锁
	defer c.mu.Unlock()

	// 返回消息列表
	return c.Messages
}

// GetMessages get messages with max window size
func (c *Conversation) GetMessages() []*schema.Message {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.Messages) > c.maxWindowSize {
		return c.Messages[len(c.Messages)-c.maxWindowSize:]
	}

	return c.Messages
}

// load函数用于加载对话文件
func (c *Conversation) Load() error {
	// 打开对话文件
	reader, err := os.Open(c.filePath)
	if err != nil {
		// 如果打开文件失败，返回错误
		return fmt.Errorf("failed to open file: %w", err)
	}
	// 关闭文件
	defer reader.Close()

	// 创建一个扫描器，用于逐行读取文件内容
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		// 读取一行内容
		line := scanner.Text()
		// 定义一个消息结构体
		var msg schema.Message
		// 将读取的内容解析为消息结构体
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// 如果解析失败，返回错误
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}
		// 将解析后的消息添加到对话的消息列表中
		c.Messages = append(c.Messages, &msg)
	}

	// 检查扫描器是否有错误
	if err := scanner.Err(); err != nil {
		// 如果有错误，返回错误
		return fmt.Errorf("scanner error: %w", err)
	}

	// 返回nil表示加载成功
	return nil
}

// 保存消息到文件
func (c *Conversation) Save(msg *schema.Message) {
	// 将消息转换为JSON字符串
	str, _ := json.Marshal(msg)

	// Append to file
	f, err := os.OpenFile(c.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(str)
	f.WriteString("\n")
}
