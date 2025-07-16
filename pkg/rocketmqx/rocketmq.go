package rocketmqx

import (
	"context"
	"errors"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"go.uber.org/zap"
	"sync"
)

// RocketMQ ...
type RocketMQ struct {
	Name                string
	NameSrvs            []string
	Brokers             []string
	Consumers           map[string]rocketmq.PushConsumer
	Producers           map[string]rocketmq.Producer
	mux                 sync.RWMutex
	Credentials         *primitive.Credentials
	consumerGroupPrefix string
	log                 *zap.SugaredLogger
}

// New 创建客户端实例
func NewRocketMQ(name string, o *RocketMQOptions, log *zap.SugaredLogger) *RocketMQ {
	if len(o.ConsumerGroupPrefix) == 0 {
		panic("消费者群组名称前缀不能为空， 请检查RocketMQ配置项进行配置")
	}

	r := &RocketMQ{
		Name:                name,
		NameSrvs:            o.NameSrvs,
		Brokers:             o.Brokers,
		Credentials:         o.Credentials,
		Consumers:           map[string]rocketmq.PushConsumer{},
		Producers:           map[string]rocketmq.Producer{},
		consumerGroupPrefix: o.ConsumerGroupPrefix,
		log:                 log,
	}
	// 使用自定义的日志记录器
	rlog.SetLogger(r.NewLog(name, log, o.LogLevel))
	return r
}

// CreateTopic 创建主题
func (r *RocketMQ) CreateTopic(ctx context.Context, topicName string, nameSrvs, brokers []string) error {
	testAdmin, err := admin.NewAdmin(admin.WithResolver(primitive.NewPassthroughResolver(nameSrvs)))
	if err != nil {
		r.log.Errorf("RocketMq[%s],连接失败: %s", r.Name, err.Error())
		return errors.New(fmt.Sprintf("RocketMq[%s],连接失败: %s", r.Name, err))
	}

	for _, broker := range brokers {
		err = testAdmin.CreateTopic(ctx, admin.WithTopicCreate(topicName), admin.WithBrokerAddrCreate(broker))
		if err != nil {
			r.log.Errorf("RocketMq[%s],broker:%s 创建主题 %s 失败: %s", r.Name, broker, topicName, err.Error())
		} else {
			r.log.Debugf("RocketMq[%s],broker:%s 创建主题: %s 成功", r.Name, broker, topicName)
		}
	}
	return nil
}

// BatchCreateTopic 批量创建主题
func (r *RocketMQ) BatchCreateTopic(ctx context.Context, topicNames []string, nameSrvs, brokers []string) error {
	testAdmin, err := admin.NewAdmin(admin.WithResolver(primitive.NewPassthroughResolver(nameSrvs)))
	if err != nil {
		return errors.New(fmt.Sprintf("RocketMq[%s],连接失败: %s", r.Name, err))
	}
	defer testAdmin.Close()
	for _, topicName := range topicNames {
		for _, broker := range brokers {
			err = testAdmin.CreateTopic(ctx, admin.WithTopicCreate(topicName), admin.WithBrokerAddrCreate(broker))
			if err != nil {
				r.log.Errorf("RocketMq[%s],broker:%s 创建主题 %s 失败: %s", r.Name, broker, topicName, err.Error())
			} else {
				r.log.Debugf("RocketMq[%s],broker:%s 创建主题: %s 成功", r.Name, broker, topicName)
			}
		}
	}
	return nil
}

// Consumer 添加消费者
func (r *RocketMQ) Consumer(ctx context.Context, topic string, f ConsumerCallback, optFuncs ...ConsumerOptionFunc) error {
	opts := newConsumerOptions()
	for _, optionFunc := range optFuncs {
		optionFunc(opts)
	}

	// 如果没有指定分组名, 先用主题和标签表达式生成一个
	if opts.GroupName == "" {
		opts.GroupName = r.buildConsumerGroupName(topic, opts.TagExpression, r.consumerGroupPrefix)
	}

	var (
		c   rocketmq.PushConsumer
		err error
		ok  bool
	)

	// 同一分组的消费者, 如果已经存在则复用, 提高程序性能
	r.mux.Lock()
	defer r.mux.Unlock()
	if c, ok = r.Consumers[opts.GroupName]; !ok {
		c, err = r.newConsumer(ctx, opts)
		if err != nil {
			return errors.New(fmt.Sprintf("RocketMq[%s],创建消费者失败: %s", r.Name, err.Error()))
		}
		r.Consumers[opts.GroupName] = c
	}

	// 标签选择器
	tagSelector := consumer.MessageSelector{}
	if opts.TagExpression != "" {
		tagSelector = consumer.MessageSelector{
			Type:       consumer.TAG,
			Expression: opts.TagExpression,
		}
	}
	err = c.Subscribe(topic, tagSelector, f)
	if err != nil {
		return errors.New(fmt.Sprintf("RocketMq[%s],消费者订阅失败: %s", r.Name, err.Error()))
	}

	return nil
}

// PublishSync 同步发送消息
func (r *RocketMQ) PublishSync(ctx context.Context, msg *Message, optFuncs ...ProducerFunc) error {
	return r.publish(ctx, msg, false, nil, optFuncs...)
}

// PublishAsync 异步发送消息
func (r *RocketMQ) PublishAsync(ctx context.Context, msg *Message, callback SendAsyncCallback, optFuncs ...ProducerFunc) error {
	return r.publish(ctx, msg, true, callback, optFuncs...)
}

