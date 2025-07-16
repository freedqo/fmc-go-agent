package rocketmqx

import (
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap/zapcore"
)

type RocketMQOptions struct {
	// 这个地址相当于注册中心, 会有心跳检测检查集群中的broker, 然后更新进来
	NameSrvs []string `json:"name_srvs,omitempty" mapstructure:"name-srvs"`
	// 这个地址只在创建主题的时候需要
	Brokers     []string               `json:"brokers,omitempty" mapstructure:"brokers"`
	Credentials *primitive.Credentials `json:"credentials,omitempty" mapstructure:"credentials"`
	// 日志级别: debug,info,warn,error
	LogLevel zapcore.Level `json:"log_level,omitempty" mapstructure:"log-level"`
	// 消费者分组前缀
	// 该配置是为确保**不同**服务中订阅了相同主题的消费者分在不同的群组(群组名不重复)
	// 从而不同服务可以各自收到消息并独立地处理消息
	// 群组的作用是使消费者在集群中起到负载均衡的作用, 对于**集群模式**的消息, 相同群组的消费者,只要有一个消费了,同群组的消费者就不会再消费了
	// 如果不同服务的消费者有相同的群组, 就很可能会发生一条消息只被其中一个服务的消费者消费
	ConsumerGroupPrefix string `json:"consumer_group_prefix" mapstructure:"consumer-group-prefix"`
}

type RocketMQCredentials struct {
	AccessKey     string `json:"access_key,omitempty" mapstructure:"access-key"`
	SecretKey     string `json:"secret_key,omitempty" mapstructure:"secret-key"`
	SecurityToken string `json:"security_token,omitempty" mapstructure:"security-token"`
}

func NewRocketMQOptions() *RocketMQOptions {
	return &RocketMQOptions{
		NameSrvs:    []string{"127.0.0.1:9876"},
		Credentials: nil,
		LogLevel:    zapcore.WarnLevel,
	}
}

// Validate 配置项校验
func (o *RocketMQOptions) Validate() []error {
	errs := []error{}

	return errs
}

