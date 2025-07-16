package ukafka

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/tools/tls"
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
	"github.com/freedqo/fmc-go-agent/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UKafka 实现 MessageAgentIf 接口
type UKafka struct {
	Name                 string
	log                  *zap.SugaredLogger
	config               *sarama.Config
	syncProducer         sarama.SyncProducer
	asyncProducer        sarama.AsyncProducer
	consumer             sarama.ConsumerGroup
	opt                  *Options
	ctx                  context.Context
	lCtx                 context.Context
	lCancel              context.CancelFunc
	lCtx1                context.Context
	lCancel1             context.CancelFunc
	recFuncs             []*umsg.RecEventFunc
	connectState         umsg.ClientConnectState
	mu                   sync.RWMutex
	wg                   sync.WaitGroup
	receivingQueue       chan *sarama.ConsumerMessage
	sendingQueue         chan *umsg.UMsg
	publishTopicMap      map[string]*KafkaTopic
	SubscriptionTopicMap map[string]*KafkaTopic
	reconnectTicker      *time.Ticker
	reconnectAttempts    int
	maxReconnectAttempts int
	reconnectDelay       time.Duration
}

// New 创建Kafka客户端实例
func New(name string, opt *Options, log *zap.SugaredLogger) If {
	// 创建默认配置
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 5 * time.Second
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// 设置分区器
	switch opt.PartitionMode {
	case ManualPartition:
		config.Producer.Partitioner = sarama.NewManualPartitioner // 手动分区（使用消息中的Partition字段）
	case RandomPartition:
		config.Producer.Partitioner = sarama.NewRandomPartitioner // 随机分区
	case RoundRobinPartition:
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner // 轮询分区
	case HashPartition:
		config.Producer.Partitioner = sarama.NewHashPartitioner // 哈希分区（需消息包含Key）
	default:
		panic(fmt.Errorf("unsupported partition mode: %d", opt.PartitionMode))
	}
	// 设置Kafka版本
	if opt.Version != "" {
		version, err := sarama.ParseKafkaVersion(opt.Version)
		if err != nil {
			log.Fatalf("解析Kafka版本失败: %v", err)
		}
		config.Version = version
	} else {
		config.Version = sarama.V2_8_1_0 // 默认版本
	}
	// 设置SASL认证
	if opt.EnableSASL {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = opt.User
		config.Net.SASL.Password = opt.Password
		// 设置认证机制
		switch opt.SASLMechanism {
		case sarama.SASLTypePlaintext:
			config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		case sarama.SASLTypeSCRAMSHA256:
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		case sarama.SASLTypeSCRAMSHA512:
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		case sarama.SASLTypeOAuth:
			panic(fmt.Errorf("unsupported SASL mechanism: %s,err: 需额外实现OAuth2 Token获取逻辑,未实现", opt.SASLMechanism))
			// 需额外实现OAuth2 Token获取逻辑
			// config.Net.SASL.Mechanism = sarama.SASLTypeOAuth
		default:
			panic(fmt.Errorf("unsupported SASL mechanism: %s", opt.SASLMechanism))
		}
	} else {
		config.Net.SASL.Enable = false
	}
	// 设置TLS
	if opt.EnableTLS {
		tlsConfig, err := tls.NewConfig(opt.ClientCertPath, opt.ClientKeyPath)
		if err != nil {
			panic(fmt.Errorf("kafka客户端,创建TLS配置失败: %s", err.Error()))
		}
		if opt.CACertPath != "" {
			rootCAsBytes, err := os.ReadFile(opt.CACertPath)
			if err != nil {
				panic(fmt.Errorf("kafka客户端,加载 CA 证书失败: %s", err.Error()))
			}
			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(rootCAsBytes) {
				panic(fmt.Errorf("kafka客户端,加载 CA 证书失败: %s", err.Error()))
			}
			// Use specific root CA set vs the host's set
			tlsConfig.RootCAs = certPool
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Config.InsecureSkipVerify = opt.InsecureSkipVerify
	}
	k := &UKafka{
		Name:   name,
		config: config,
		log:    log,
		opt:    opt,
		connectState: umsg.ClientConnectState{
			ClientID:    name,
			IsConnected: false,
			RemoteAddr:  "",
		},
		recFuncs:             make([]*umsg.RecEventFunc, 0),
		receivingQueue:       make(chan *sarama.ConsumerMessage, 10*1024),
		sendingQueue:         make(chan *umsg.UMsg, 10*1024),
		reconnectAttempts:    0,
		maxReconnectAttempts: 10,
		reconnectDelay:       5 * time.Second,
	}
	k.publishTopicMap = make(map[string]*KafkaTopic)
	for _, v := range k.opt.PublishTopics {
		k.publishTopicMap[v.Topic] = v
	}
	k.SubscriptionTopicMap = make(map[string]*KafkaTopic)
	for _, v := range k.opt.SubscriptionTopics {
		k.SubscriptionTopicMap[v.MapTopic] = v
	}
	return k
}

func (k *UKafka) StartKafka() {
	defer func() {
		if err := recover(); err != nil {
			// 生成一个唯一的错误ID，用于后续的错误跟踪
			PanicId := uuid.New().ID()
			// 记录 panic 信息
			k.log.Errorf("推送kafka消息,遇到异常,PanicId: %d, Panic: %v", PanicId, err)
			// 打印堆栈跟踪信息（可选）
			k.log.Errorf(utils.StackSkip(1, -1))
		}
	}()
	// 等待所有goroutine完成
	k.wg.Wait()
	k.lCtx1, k.lCancel1 = context.WithCancel(k.ctx)
	if k.opt.EnableCreatTopic {
		// 创建kafka主题
		if err := k.createTopicIfNotExists(k.opt.PublishTopics); err != nil {
			k.log.Errorf("创建kafka主题失败: %v", err)
		}
	}
	once := sync.Once{}
	time.Sleep(1 * time.Second)
	// 启动消费者
	if err := k.startConsumer(&once); err != nil {
		k.log.Errorf("启动消费者失败: %v", err)
	}

	// 初始化生产者
	if err := k.initProducers(&once); err != nil {
		k.log.Errorf("初始化生产者失败: %v", err)
	}

}

// 创建kafka主题（如果不存在）
func (k *UKafka) createTopicIfNotExists(topics []*KafkaTopic) error {
	// 创建临时客户端检查kafka主题
	client, err := sarama.NewClient(k.opt.Brokers, k.config)
	if err != nil {
		return fmt.Errorf("创建临时客户端失败: %w", err)
	}
	defer client.Close()

	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		return fmt.Errorf("创建集群管理客户端失败: %w", err)
	}
	defer admin.Close()

	// 确保所有发布kafka主题存在
	for _, topic := range topics {
		topic1 := topic.Topic
		if topic.MapTopic != "" {
			topic1 = topic.MapTopic
		}

		// kafka主题配置
		topicDetail := sarama.TopicDetail{
			NumPartitions:     k.opt.NumPartitions,       // 分区数
			ReplicationFactor: int16(len(k.opt.Brokers)), // 副本因子
			ConfigEntries: map[string]*string{
				"retention.ms": toString("3600000"), // 1小时（3600秒 × 1000毫秒）
			},
			ReplicaAssignment: map[int32][]int32{},
		}
		// 创建kafka主题
		k.log.Infof("创建kafka主题: %s,配置：%v", topic1, topicDetail)
		err = admin.CreateTopic(topic1, &topicDetail, false)
		if err != nil {
			if strings.Contains(fmt.Sprintf("%s", err), sarama.ErrTopicAlreadyExists.Error()) {
				k.log.Infof("kafka主题已存在: %s", topic1)
			} else {
				return err
			}
		}
		k.log.Infof("kafka主题创建成功: %s", topic1)
	}
	// 检查kafka主题是否已存在
	existingTopics1, err := admin.ListTopics()
	if err != nil {
		return fmt.Errorf("获取kafka主题列表失败: %w", err)
	}
	for _, v := range topics {
		topic1 := v.Topic
		if v.MapTopic != "" {
			topic1 = v.MapTopic
		}
		if _, exists := existingTopics1[topic1]; !exists {
			return fmt.Errorf("期望的kafka主题[%s]不存在", topic1)
		} else {
			k.log.Infof("kafka主题[%s]存在,配置：%v", topic1, existingTopics1[topic1])
		}
	}
	for topic, opt := range existingTopics1 {
		isexit := false
		for _, v := range topics {
			topic1 := v.Topic
			if v.MapTopic != "" {
				topic1 = v.MapTopic
			}
			if topic == topic1 {
				isexit = true
			}
		}
		if isexit && opt.NumPartitions != k.opt.NumPartitions {
			err := admin.CreatePartitions(topic, k.opt.NumPartitions, nil, false)
			if err != nil {
				k.log.Errorf("创建kafka主题[%s]分区失败: %s", topic, err.Error())
				return err
			}
			k.log.Infof("创建kafka主题[%s]分区数量[%d]成功", topic, k.opt.NumPartitions)
		}
	}

	return nil
}

// 辅助函数：将字符串转为指针
func toString(s string) *string {
	return &s
}

// 初始化生产者
func (k *UKafka) initProducers(once *sync.Once) error {
	var err error

	// 创建同步生产者
	k.syncProducer, err = sarama.NewSyncProducer(k.opt.Brokers, k.config)
	if err != nil {
		return fmt.Errorf("创建kafka同步生产者失败: %w", err)
	}

	// 创建异步生产者
	k.asyncProducer, err = sarama.NewAsyncProducer(k.opt.Brokers, k.config)
	if err != nil {
		return fmt.Errorf("创建kafka异步生产者失败: %w", err)
	}
	k.wg.Add(1)
	go func() {
		defer func() {
			k.wg.Done()
			k.log.Info("kafka异步消息发送协程退出")
			k.connectState.IsConnected = false
			once.Do(func() {
				k.log.Info("kafka异步消息发送协程，请求关闭lCancel1")
				k.lCancel1()
			})
		}()
		defer func() {
			if k.asyncProducer != nil {
				k.asyncProducer.Close()
			}
			if k.syncProducer != nil {
				k.syncProducer.Close()
			}
		}()
		k.log.Infof("kafka异步消息发送协程已启动")
		for {
			select {
			case <-k.lCtx1.Done():
				return
			case msg, ok := <-k.asyncProducer.Successes():
				{
					if !ok {
						k.log.Debugf("kafka异步消息发送成功: kafka主题=%s, 分区=%d, 偏移量=%d",
							msg.Topic, msg.Partition, msg.Offset)
					}
				}
			case msg, ok := <-k.asyncProducer.Errors():
				{
					if !ok {
						k.log.Errorf("kafka异步生产者错误: %v", msg)
					}
				}
			}
		}
	}()
	return nil
}

// 启动消费者
func (k *UKafka) startConsumer(once *sync.Once) error {
	var err error
	// 创建消费者组
	k.consumer, err = sarama.NewConsumerGroup(k.opt.Brokers, k.opt.GroupID, k.config)
	if err != nil {
		return fmt.Errorf("创建kafka消费者组失败: %w", err)
	}

	// 获取订阅kafka主题
	topics := make([]string, 0)
	for _, v := range k.opt.SubscriptionTopics {
		topic := v.Topic
		if strings.TrimSpace(v.MapTopic) != "" {
			topic = v.MapTopic
		}
		topics = append(topics, topic)
	}

	// 启动消费协程
	k.wg.Add(1)
	go func() {
		defer func() {
			k.wg.Done()
			if r := recover(); r != nil {
				// 生成一个唯一的错误ID，用于后续的错误跟踪
				PanicId := uuid.New().ID()
				// 记录 panic 信息
				k.log.Errorf("kafka消费者组消费异常,遇到异常,PanicId: %d, Panic: %v", PanicId, err)
				// 打印堆栈跟踪信息（可选）
				k.log.Errorf(utils.StackSkip(1, -1))
			}
			k.log.Info("kafka消费者组消费协程退出")
			k.connectState.IsConnected = false
			once.Do(func() {
				k.log.Info("kafka消费者组消费协程，请求关闭lCancel1")
				k.lCancel1()
			})
		}()
		defer k.consumer.Close()
		k.log.Infof("kafka消费者组开始消费: %s", k.opt.GroupID)
		// 开始消费
		err := k.consumer.Consume(k.lCtx1, topics, k)
		if err != nil {
			k.log.Errorf("kafka消费者组消费错误: %v", err)
		}
	}()

	return nil
}

// Start 启动客户端监控
func (k *UKafka) Start(ctx context.Context) (<-chan struct{}, error) {
	k.ctx = ctx
	k.lCtx, k.lCancel = context.WithCancel(ctx)
	once := sync.Once{}
	doneCh := make(chan struct{})
	k.log.Infof("开始启动kafka消息服务")
	testClient, err := sarama.NewClient(k.opt.Brokers, k.config)
	if err != nil {
		k.connectState.IsConnected = false
		k.log.Errorf("连接Kafka失败: %v,%v", k.opt.Brokers, err)
	} else {
		k.connectState.IsConnected = true
		k.log.Infof("连接Kafka成功: %v", k.opt.Brokers)
	}
	if testClient != nil {
		defer testClient.Close()
	}

	if k.connectState.IsConnected {
		k.StartKafka()
	}

	// 构建连接状态监听协程
	k.reconnectTicker = time.NewTicker(20 * time.Second)
	// 构建消息队列和通信状态监视协程
	go func() {
		defer func() {
			once.Do(func() {
				close(doneCh)
			})
		}()
		defer func() {
			if err := recover(); err != nil {
				// 生成一个唯一的错误ID，用于后续的错误跟踪
				PanicId := uuid.New().ID()
				// 记录 panic 信息
				k.log.Errorf("ukafka消息队列和通信状态监视协程,遇到异常,PanicId: %d, Panic: %v", PanicId, err)
				// 打印堆栈跟踪信息（可选）
				k.log.Errorf(utils.StackSkip(1, -1))
			}
		}()
		for {
			select {
			case <-k.lCtx.Done():
				{
					return
				}
			case <-k.reconnectTicker.C:
				{
					if k.connectState.IsConnected {
						break
					}
					// 检测连接状态
					testClient1, err := sarama.NewClient(k.opt.Brokers, k.config)
					if err != nil {
						k.connectState.IsConnected = false
						k.log.Errorf("连接Kafka失败: %v,%v", k.opt.Brokers, err)
						break
					}
					k.log.Infof("连接Kafka成功: %v", k.opt.Brokers)
					testClient1.Close()
					k.connectState.IsConnected = true
					if k.connectState.IsConnected {
						k.StartKafka()
					}
					break
				}
			case msg, ok := <-k.receivingQueue:
				{
					if ok {
						k.handleReciveMsg(msg)
					}
				}
			case msg, ok := <-k.sendingQueue:
				{
					if ok {
						k.handleSendMsg(msg)
					}
				}

			}
		}
	}()
	k.log.Infof("kafka消息服务启动完成")
	return doneCh, nil
}

// Stop 停止客户端监控
func (k *UKafka) Stop() error {
	if k.lCancel != nil {
		k.lCancel()
	}
	if k.reconnectTicker != nil {
		k.reconnectTicker.Stop()
	}

	// 关闭消费者组
	if k.consumer != nil {
		if err := k.consumer.Close(); err != nil {
			k.log.Errorf("关闭kafka消费者组失败: %v", err)
		}
	}
	// 关闭生产者
	if k.syncProducer != nil {
		if err := k.syncProducer.Close(); err != nil {
			k.log.Errorf("关闭kafka同步生产者失败: %v", err)
		}
	}
	if k.asyncProducer != nil {
		if err := k.asyncProducer.Close(); err != nil {
			k.log.Errorf("关闭kafka异步生产者失败: %v", err)
		}
	}

	// 更新连接状态
	k.connectState.IsConnected = false

	return nil
}

// RestStart 重启客户端监控
func (k *UKafka) RestStart() (<-chan struct{}, error) {
	if err := k.Stop(); err != nil {
		return nil, fmt.Errorf("重启前停止失败: %w", err)
	}
	return k.Start(context.Background())
}

// 构建Kafka生产者消息
func (k *UKafka) buildKafkaMessage(msg *umsg.UMsg) (*sarama.ProducerMessage, error) {
	// 将消息转换为JSON格式
	dataJson, err := json.Marshal(msg.Msg.OperateData)
	if err != nil {
		return nil, err
	}
	// 确保消息大小不超过Kafka集群配置的最大消息大小
	if len(dataJson) > 1000000 { // 默认Kafka最大消息大小为1MB
		return nil, fmt.Errorf("消息大小超过1MB: %d 字节", len(dataJson))
	}
	// 创建消息
	topic := msg.Msg.Operate
	partition := k.opt.PublishPartition

	kTopic, ok := k.publishTopicMap[msg.Msg.Operate]
	if !ok {
		return nil, fmt.Errorf("未找到对应的Topic,topic:%s", msg.Msg.Operate)
	}
	if strings.TrimSpace(kTopic.MapTopic) != "" {
		topic = kTopic.MapTopic
	}
	if kTopic.Partition > -1 {
		partition = kTopic.Partition
	}

	kMsg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: partition,
		Key:       nil,
		Value:     sarama.StringEncoder(dataJson),
		Timestamp: time.Now(),
	}
	return kMsg, nil
}

func (k *UKafka) handleSendMsg(msg *umsg.UMsg) {
	// 检查连接状态
	k.mu.RLock()
	connected := k.connectState.IsConnected
	k.mu.RUnlock()

	if !connected {
		k.log.Warnf("Kafka连接断开，消息将被丢弃: %v", msg.Msg.Operate)
		return
	}

	// 创建Kafka消息
	kafkaMsg, err := k.buildKafkaMessage(msg)
	if err != nil {
		k.log.Errorf("构建Kafka消息失败: %v", err)
		return
	}

	// 异步或同步发送
	if msg.IsASync {
		k.asyncProducer.Input() <- kafkaMsg
		k.log.Debugf("Kafka[%s],异步发送消息:%s", k.Name, msg.String("发送"))
	} else {
		p, o, err := k.syncProducer.SendMessage(kafkaMsg)
		if err != nil {
			k.log.Errorf("同步发送Kafka消息失败: %v,消息：%v", err, kafkaMsg)
		} else {
			k.log.Debugf("Kafka[%s],同步发送消息:%s,分区地址：%d,偏移量：%d", k.Name, msg.String("发送"), p, o)
		}
	}
}

// Publish 推送消息到Kafka
func (k *UKafka) Publish(msg *umsg.UMsg) {
	defer func() {
		if err := recover(); err != nil {
			// 生成一个唯一的错误ID，用于后续的错误跟踪
			PanicId := uuid.New().ID()
			// 记录 panic 信息
			k.log.Errorf("推送kafka消息,遇到异常,PanicId: %d, Panic: %v", PanicId, err)
			// 打印堆栈跟踪信息（可选）
			k.log.Errorf(utils.StackSkip(1, -1))
		}
	}()
	k.sendingQueue <- msg

}

// SubscribeRecEvent 注册接收消息的回调函数
func (k *UKafka) SubscribeRecEvent(fun *umsg.RecEventFunc) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.recFuncs = append(k.recFuncs, fun)
}

// UnSubscribeRecEvent 注销接收消息的回调函数
func (k *UKafka) UnSubscribeRecEvent(fun *umsg.RecEventFunc) {
	k.mu.Lock()
	defer k.mu.Unlock()
	for i, f := range k.recFuncs {
		if f == fun {
			k.recFuncs = append(k.recFuncs[:i], k.recFuncs[i+1:]...)
			break
		}
	}
}

// GetConnectState 获取连接状态
func (k *UKafka) GetConnectState() umsg.ClientConnectState {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.connectState
}

func (k *UKafka) handleReciveMsg(msg *sarama.ConsumerMessage) {
	defer func() {
		if err := recover(); err != nil {
			// 生成一个唯一的错误ID，用于后续的错误跟踪
			PanicId := uuid.New().ID()
			// 记录 panic 信息
			k.log.Errorf("处理待推送消息,遇到异常,PanicId: %d, Panic: %v", PanicId, err)
			// 打印堆栈跟踪信息（可选）
			k.log.Errorf(utils.StackSkip(1, -1))
		}
	}()
	uMsg := umsg.NewUMsg(&umsg.Message{
		MessageBase: umsg.MessageBase{
			ClientID:        k.Name,
			Operate:         msg.Topic,
			IsReplyOperate:  false,
			OperateID:       fmt.Sprintf("%s%s%d%s", msg.Topic, msg.Key, msg.Partition, strconv.FormatInt(msg.Offset, 10)),
			OperateDataType: fmt.Sprintf("%T", msg),
		},
		OperateData: msg,
	}, k.Name, nil)
	k.log.Debugf("Kafka[%s],收到消息:%s", k.Name, uMsg.String("接收"))
	uMsg.Msg.ClientID = k.Name
	uMsg.Flag = k.Name
	for _, item := range k.recFuncs {
		(*item)(uMsg)
	}
}

func (k *UKafka) Setup(session sarama.ConsumerGroupSession) error {
	k.log.Infof("Kafka[%s],开始消费会话,消费组ID：%s", k.Name, session.MemberID())
	return nil
}

func (k *UKafka) Cleanup(session sarama.ConsumerGroupSession) error {
	k.log.Warnf("Kafka[%s],结束消费会话,消费组ID：%s", k.Name, session.MemberID())
	return nil
}

func (k *UKafka) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	defer func() {
		if err := recover(); err != nil {
			// 生成一个唯一的错误ID，用于后续的错误跟踪
			PanicId := uuid.New().ID()
			// 记录 panic 信息
			k.log.Errorf("处理待推送消息,遇到异常,PanicId: %d, Panic: %v", PanicId, err)
			// 打印堆栈跟踪信息（可选）
			k.log.Errorf(utils.StackSkip(1, -1))
		}
	}()
	k.log.Infof("Kafka[%s],开始消费进入循环消费会话,消费组ID：%s", k.Name, session.MemberID())

	// 消费消息的主循环
	for message := range claim.Messages() {
		v, ok := k.SubscriptionTopicMap[message.Topic]
		if !ok {
			k.log.Warnf("Kafka[%s] 消费消息kafka主题未注册: %s,Msg:%v", k.Name, message.Topic, message)
			// 标记消息已消费（自动提交时会定期提交偏移量）
			session.MarkMessage(message, "")
			continue
		}
		if v.Topic == "Println" {
			msg, err := json.Marshal(message)
			if err == nil {
				k.log.Infof("Kafka[%s] 消费仅打印的消息,Msg:%s,value:%s", k.Name, msg, string(message.Value))
			}
			// 标记消息已消费（自动提交时会定期提交偏移量）
			session.MarkMessage(message, "")
			continue
		}
		if v.Partition != message.Partition {
			k.log.Warnf("Kafka[%s] 消费消息kafka分区未注册: %d,Msg:%v", k.Name, message.Partition, message)
			continue
		}
		// 使用队列，避免高并发引起的资源竞争
		k.receivingQueue <- message
		// 标记消息已消费（自动提交时会定期提交偏移量）
		session.MarkMessage(message, "")
	}
	return nil
}

var _ If = &UKafka{}
