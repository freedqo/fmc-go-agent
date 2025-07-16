package rocketmqx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type ProducerOptions struct {
	GroupName  string
	RetryTimes int
}

type ProducerFunc func(options *ProducerOptions)

func newProducerOptions() *ProducerOptions {
	return &ProducerOptions{
		GroupName:  "",
		RetryTimes: 3,
	}
}

func WithProducerGroupName(name string) ProducerFunc {
	return func(opt *ProducerOptions) {
		opt.GroupName = name
	}
}

func WithProducerRetryTimes(retryTimes int) ProducerFunc {
	return func(opt *ProducerOptions) {
		opt.RetryTimes = retryTimes
	}
}

// SendAsyncCallback 异步发送回调方法类型
type SendAsyncCallback func(ctx context.Context, result *primitive.SendResult, e error)

func (r *RocketMQ) publish(ctx context.Context, msg *Message, isAsync bool, callback SendAsyncCallback, optFuncs ...ProducerFunc) error {
	opt := newProducerOptions()
	opt.GroupName = r.buildProducerGroupName(msg)

	if len(optFuncs) > 0 {
		for _, f := range optFuncs {
			f(opt)
		}
	}

	var (
		p   rocketmq.Producer
		err error
		ok  bool
	)

	// 同一分组的生产者实例进行复用, 提高性能
	r.mux.Lock()
	if p, ok = r.Producers[opt.GroupName]; !ok {
		p, err = r.newProducer(ctx, opt)
		if err != nil {
			r.mux.Unlock()
			return errors.New(fmt.Sprintf("创建生产者失败: %s", opt.GroupName))
		}
		err = p.Start()
		if err != nil {
			r.mux.Unlock()
			return errors.New(fmt.Sprintf("生产者启动失败: %s", err))
		}
		r.Producers[opt.GroupName] = p
	}
	r.mux.Unlock()

	rMsgList, err := r.buildMessage(msg)
	if err != nil {
		return err
	}

	if isAsync { // 异步发送
		err = p.SendAsync(ctx, callback, rMsgList...)

		if err != nil {
			return errors.New(fmt.Sprintf("异步发送消息失败: %s", err))
		}
	} else { // 同步发送
		res, err := p.SendSync(ctx, rMsgList...)

		if err != nil {
			return errors.New(fmt.Sprintf("发送消息失败: %s", err))
		}
		if res.Status != primitive.SendOK {
			return errors.New(fmt.Sprintf("发送消息失败, 消息错误码: %d", res.Status))
		}
	}

	return nil
}

func (r *RocketMQ) newProducer(ctx context.Context, opt *ProducerOptions) (rocketmq.Producer, error) {
	if r.Credentials != nil {
		return rocketmq.NewProducer(
			producer.WithNsResolver(primitive.NewPassthroughResolver(r.NameSrvs)),
			producer.WithGroupName(opt.GroupName),
			producer.WithRetry(opt.RetryTimes),
			producer.WithCredentials(*r.Credentials),
		)
	}
	return rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver(r.NameSrvs)),
		producer.WithGroupName(opt.GroupName),
		producer.WithRetry(opt.RetryTimes),
	)
}

func (r *RocketMQ) buildProducerGroupName(msg *Message) string {
	group := msg.Topic
	if msg.GetTag() != "" {
		group = group + "-" + msg.GetTag()
	}

	return group
}

func (r *RocketMQ) buildMessage(msg *Message) ([]*primitive.Message, error) {
	if msg.Topic == "" {
		return nil, errors.New("主题名称不能为空")
	}

	if len(msg.BodyList) == 0 {
		return nil, errors.New("消息内容不能为空")
	}

	rMsgList := make([]*primitive.Message, 0)

	for _, item := range msg.BodyList {
		var b []byte

		if v, ok := item.(string); ok {
			b = []byte(v)
		} else if vb, ok := item.([]byte); ok {
			b = vb
		} else {
			enc, err := json.Marshal(item)
			if err != nil {
				return nil, errors.New("序列化消息内容错误")
			}
			b = enc
		}

		rMsg := primitive.NewMessage(msg.Topic, b)
		if msg.Tag != "" {
			rMsg.WithTag(msg.Tag)
		}
		if msg.DelayLevel > 0 {
			rMsg.WithDelayTimeLevel(msg.DelayLevel)
		}

		if len(msg.Keys) > 0 {
			rMsg.WithKeys(msg.Keys)
		}

		rMsgList = append(rMsgList, rMsg)
	}

	return rMsgList, nil
}

