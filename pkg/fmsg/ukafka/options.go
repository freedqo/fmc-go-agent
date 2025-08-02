package ukafka

func NewDefaultOptions() *Options {
	SubscriptionTopics := make([]*KafkaTopic, 0)
	SubscriptionTopics = append(SubscriptionTopics, &KafkaTopic{
		Topic:     "test12",
		MapTopic:  "",
		Partition: 14,
	})
	PublishTopics := make([]*KafkaTopic, 0)
	PublishTopics = append(PublishTopics, &KafkaTopic{
		Topic:     "test12",
		MapTopic:  "",
		Partition: 14,
	})

	return &Options{
		Brokers:               []string{"localhost:9092"},
		GroupID:               "",
		User:                  "",
		Password:              "",
		Version:               "",
		EnableCreatTopic:      false,
		EnableSASL:            false,
		SASLMechanism:         "",
		EnableTLS:             false,
		CACertPath:            "",
		InsecureSkipVerify:    false,
		ClientCertPath:        "",
		ClientKeyPath:         "",
		SubscriptionTopics:    SubscriptionTopics,
		SubscriptionPartition: 0,
		PublishTopics:         PublishTopics,
		PublishPartition:      0,
		NumPartitions:         16,
	}
}

type Options struct {
	Brokers               []string      `comment:"Broker地址列表"`                                                         // Kafka broker地址列表
	GroupID               string        `comment:"消费者组ID"`                                                             // 消费者组ID
	Version               string        `comment:"Kafka版本号,例如:0.2.6.0"`                                                // Kafka版本号
	EnableCreatTopic      bool          `comment:"是否创建主题"`                                                             // 是否创建主题
	EnableSASL            bool          `comment:"是否启用SASL认证"`                                                         // 是否启用SASL认证
	SASLMechanism         string        `comment:"SASL认证机制,可选值:PLAIN,SCRAM-SHA-256,SCRAM-SHA-512"`                     // SASL认证机制
	User                  string        `comment:"用户名"`                                                                // 用户名
	Password              string        `comment:"密码"`                                                                 // 密码
	EnableTLS             bool          `comment:"是否启用TLS"`                                                            // 是否启用TLS
	CACertPath            string        `comment:"CA证书路径"`                                                             // CA证书路径
	InsecureSkipVerify    bool          `comment:"是否跳过TLS证书验证"`                                                        // 是否跳过TLS证书验证
	ClientCertPath        string        `comment:"客户端证书路径"`                                                            // 客户端证书路径
	ClientKeyPath         string        `comment:"客户端私钥路径"`                                                            // 客户端私钥路径
	NumPartitions         int32         `comment:"分区总数量"`                                                              // 分区总数量
	PartitionMode         int           `comment:"分区模式,可选值:0->Manual,手动,1:Random->随机,2:RoundRobin->轮询,3:Hash->key值哈希"` // 分区模式
	SubscriptionTopics    []*KafkaTopic `comment:"订阅的主题列表"`                                                            // 订阅的主题列表
	SubscriptionPartition int32         `comment:"默认消费分区(不能为空)"`                                                       // 生产与消费分区
	PublishTopics         []*KafkaTopic `comment:"发布的主题列表"`                                                            // 发布的主题列表
	PublishPartition      int32         `comment:"默认生产分区(不能为空)"`                                                       // 生产与消费分区
}

type KafkaTopic struct {
	Topic     string `comment:"内部定义主题"`
	MapTopic  string `comment:"外部定义主题"`
	Partition int32  `comment:"应用分区（-1的时候，使用默认分区）"`
}

const (
	ManualPartition     int = iota // 手动指定分区（使用消息中的Partition字段）
	RandomPartition                // 随机分区
	RoundRobinPartition            // 轮询分区
	HashPartition                  // 哈希分区（根据消息Key计算分区）
)
