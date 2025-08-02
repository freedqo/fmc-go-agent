package rocketmqx

import "sync"

// Message 消息定义
type Message struct {
	Topic      string
	Tag        string
	Keys       []string // 一般用于消息在业务层面的唯一标识
	BodyList   []interface{}
	DelayLevel int
	mux        sync.Mutex
}

type optionFunc func(m *Message)

// WithTag 指定消息标签
func WithTag(tag string) optionFunc {
	return func(m *Message) {
		m.SetTag(tag)
	}
}

// WithDelayLevel 指定延迟时间
func WithDelayLevel(level int) optionFunc {
	return func(m *Message) {
		m.SetDelayLevel(level)
	}
}

// WithBody 添加消息内容, 内容可以是map, 结构体, 字符串, []byte
func WithBody(body interface{}) optionFunc {
	return func(m *Message) {
		m.AddBody(body)
	}
}

func WithKey(key string) optionFunc {
	return func(m *Message) {
		m.AddKey(key)
	}
}

// NewMessage 创建消息
func NewMessage(topic string, optFuncs ...optionFunc) *Message {
	m := &Message{Topic: topic}
	if len(optFuncs) > 0 {
		for _, f := range optFuncs {
			f(m)
		}
	}

	return m
}

func (m *Message) GetTopic() string {
	return m.Topic
}

func (m *Message) GetTag() string {
	return m.Tag
}

func (m *Message) SetTopic(topic string) {
	m.Topic = topic
}

func (m *Message) SetTag(tag string) {
	if tag != "" {
		m.Tag = tag
	}
}

func (m *Message) GetDelayLevel() int {
	return m.DelayLevel
}

func (m *Message) SetDelayLevel(level int) {
	m.DelayLevel = level
}

// AddBody 添加消息内容, 内容可以是map, 结构体, 字符串, []byte
func (m *Message) AddBody(body interface{}) {
	if body == nil {
		return
	}

	m.mux.Lock()
	defer m.mux.Unlock()
	if len(m.BodyList) == 0 {
		m.BodyList = make([]interface{}, 0)
	}
	m.BodyList = append(m.BodyList, body)
}

func (m *Message) AddKey(key string) {
	if key == "" {
		return
	}

	m.mux.Lock()
	defer m.mux.Unlock()

	if len(m.Keys) == 0 {
		m.Keys = make([]string, 0)
	}
	m.Keys = append(m.Keys, key)
}
