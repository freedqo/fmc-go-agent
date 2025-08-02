package fmsg

import (
	"encoding/json"
	"fmt"
	"time"
)

type PublishFunc func(msg *UMsg)

type RecEventFunc func(msg *UMsg)

type HandlerInMsgFun func(push PublishFunc, msg *UMsg)

const (
	ToMqtt     = "ToMqtt"     // 发送到mqtt
	ToRocketMq = "ToRocketMq" // 发送到rocketmq
	ToWsServer = "ToWsServer" // 发送到ws服务端  默认输出器
	ToWsClient = "ToWsClient" // 发送到ws客户端
	ToKafKa    = "ToKafKa"    // 发送到kafka
)

func NewUMsg(topic string, msg interface{}, source string, outType []string, option interface{}) *UMsg {
	return &UMsg{
		inTimer: time.Now(),
		Topic:   topic,
		Msg:     msg,
		Sour:    source,
		OutType: outType,
		Option:  option,
	}
}

type UMsg struct {
	inTimer  time.Time   // 输入时间
	Topic    string      // 消息主题
	Msg      interface{} // 消息内容
	endTimer time.Time   // 结束时间
	Sour     string      // 上下文标记
	OutType  []string    //输出方法,ToMqtt、ToRocketMq、ToWsServer、ToWsClient
	Option   interface{} // 应用配置，针对不同的消息平台，有不一样的配置参数
}

func (m *UMsg) String(ctx string) (result string) {
	defer func() {
		if err := recover(); err != nil {
			result = fmt.Sprintf("上下文标记：%s,打印消息内容,遇到异常：%s", ctx, err)
		}
	}()
	if m.Msg == nil {
		return fmt.Sprintf("上下文标记：%s,消息内容为空", ctx)
	}
	buf := make([]byte, 0)
	buf, ok := m.Msg.([]byte)
	if !ok {
		msgStr, err := json.Marshal(m.Msg)
		if err != nil {
			return fmt.Sprintf("解码消息内容异常：%s", err.Error())
		}
		buf = msgStr
	}
	m.endTimer = time.Now()
	subTimer := time.Since(m.inTimer)
	tps := calculateThroughput(1, subTimer.Milliseconds())
	result = fmt.Sprintf(
		"上下文标记：%s,创建标志：%s,输入时间：%s,截至时间：%s,耗时：%s ,处理速度： %.2f tps(事件数/秒),消息Base：%s",
		ctx,
		m.Sour,
		m.inTimer,
		m.endTimer,
		subTimer,
		tps,
		string(buf),
	)
	return result
}

// calculateThroughput 函数用于计算吞吐量
func calculateThroughput(eventCount int, elapsedTimeMs int64) float64 {
	// 将耗时从毫秒转换为秒
	elapsedTimeSec := float64(elapsedTimeMs) / 1000000
	// 计算吞吐量，单位为事件数/秒
	throughput := float64(eventCount) / elapsedTimeSec
	return throughput
}

type TConnStatus struct {
	Name  string
	State bool
}
