package rocketmqx

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"regexp"
	"strings"
)

// ConsumerOptions 消费者配置
type ConsumerOptions struct {
	GroupName           string                // 分组名
	TagExpression       string                // 标签表达式, 比如tagA, tagA || tagB, *
	MaxReconsumeTimes   int32                 // 消费失败重试次数, rocketmq中默认为16
	ConsumerModel       consumer.MessageModel // 消费模式: 集群或广播
	ConsumerGroupPrefix string                // 消费者群组名前缀
	ConsumeFromWhere    consumer.ConsumeFromWhere
}

type ConsumerOptionFunc func(options *ConsumerOptions)

// ConsumerCallback 消费者回调方法定义
type ConsumerCallback func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error)

func newConsumerOptions() *ConsumerOptions {
	return &ConsumerOptions{
		GroupName:           "",
		MaxReconsumeTimes:   16,
		ConsumerModel:       consumer.Clustering,
		ConsumerGroupPrefix: "",
	}
}

// WithConsumerGroupName 指定分组名称
func WithConsumerGroupName(name string) ConsumerOptionFunc {
	return func(opt *ConsumerOptions) {
		opt.GroupName = name
	}
}

// WithTagExpression 指定标签表达式
func WithTagExpression(exp string) ConsumerOptionFunc {
	return func(opt *ConsumerOptions) {
		opt.TagExpression = exp
	}
}

// WithConsumerMaxReconsumeTimes 指定最大重新消费次数
func WithConsumerMaxReconsumeTimes(times int32) ConsumerOptionFunc {
	return func(opt *ConsumerOptions) {
		opt.MaxReconsumeTimes = times
	}
}

func WithConsumeFromWhere(where consumer.ConsumeFromWhere) ConsumerOptionFunc {
	return func(opt *ConsumerOptions) {
		opt.ConsumeFromWhere = where
	}
}

// WithConsumerModel 指定是集群消费还是广播消费, 不指定默认为集群消费
func WithConsumerModel(model consumer.MessageModel) ConsumerOptionFunc {
	return func(opt *ConsumerOptions) {
		opt.ConsumerModel = model
	}
}

func (r *RocketMQ) newConsumer(ctx context.Context, opt *ConsumerOptions) (rocketmq.PushConsumer, error) {
	if r.Credentials != nil {
		return rocketmq.NewPushConsumer(
			consumer.WithGroupName(opt.GroupName),
			consumer.WithNsResolver(primitive.NewPassthroughResolver(r.NameSrvs)),
			consumer.WithConsumerModel(opt.ConsumerModel),
			consumer.WithMaxReconsumeTimes(opt.MaxReconsumeTimes),
			consumer.WithCredentials(*r.Credentials),
			consumer.WithConsumeFromWhere(opt.ConsumeFromWhere),
		)
	}
	return rocketmq.NewPushConsumer(
		consumer.WithGroupName(opt.GroupName),
		consumer.WithNsResolver(primitive.NewPassthroughResolver(r.NameSrvs)),
		consumer.WithConsumerModel(opt.ConsumerModel),
		consumer.WithMaxReconsumeTimes(opt.MaxReconsumeTimes),
		consumer.WithConsumeFromWhere(opt.ConsumeFromWhere),
	)
}

// 构建消费者分组名
func (r *RocketMQ) buildConsumerGroupName(topic, tagExpression, prefix string) string {
	group := topic

	if tagExpression != "" {
		strTemp := strings.ReplaceAll(tagExpression, "*", "all-tag")
		strTemp = strings.ReplaceAll(strTemp, "||", "-OR-")
		strTemp = strings.ReplaceAll(strTemp, "&", "-AND-")
		strTemp = strings.ReplaceAll(strTemp, " ", "-")
		strTemp = regexp.MustCompile(`[^a-zA-Z0-9\- ]+`).ReplaceAllString(strTemp, "-")
		group = group + "-" + strTemp
	}

	return prefix + "-" + group
}
