package umsg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	mqtt "github.com/eclipse/paho.mqtt.golang"
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

func NewUMsg(msg *Message, Flag string, outType []string) *UMsg {
	return &UMsg{
		inTimer: time.Now(),
		Msg:     msg,
		Flag:    Flag,
		IsASync: false,
		OutType: outType,
	}
}

type UMsg struct {
	inTimer  time.Time
	Msg      *Message
	endTimer time.Time
	Flag     string
	IsASync  bool
	OutType  []string //输出方法,ToMqtt、ToRocketMq、ToWsServer、ToWsClient
}

func (m *UMsg) String(ctx string) (result string) {
	defer func() {
		if err := recover(); err != nil {
			result = fmt.Sprintf("上下文标记：%s,打印消息内容,遇到异常：%s", ctx, err)
		}
	}()
	msgBaseStr, err := json.Marshal(m.Msg.MessageBase)
	if err != nil {
		return fmt.Sprintf("解码消息内容异常：%s", err.Error())
	}
	msgBodyStr := make([]byte, 0)
	switch m.Msg.OperateDataType {
	case TransparentTransport_package_message_OperateDataType:
		{
			body, isok := m.Msg.OperateData.([]byte)
			if isok {
				msgBodyStr = body
			} else {
				body1, isok1 := m.Msg.OperateData.(string)
				if isok1 {
					decodedData, err1 := base64.StdEncoding.DecodeString(body1)
					if err1 == nil {
						msgBodyStr = decodedData
					} else {
						msgBodyStr = []byte("消息Body解码异常")
					}
				} else {
					msgBodyStr = []byte("消息Body解码异常")
				}
			}
			break
		}
	case "*mqtt.Message":
		{
			body, ok := m.Msg.OperateData.(*mqtt.Message)
			if ok {
				msgBodyStr = (*body).Payload()
			} else {
				return fmt.Sprintf("解码消息内容异常，消息类型：%s", m.Msg.OperateDataType)
			}
			break
		}
	case "*primitive.MessageExt":
		{
			body, ok := m.Msg.OperateData.(*primitive.MessageExt)
			if ok {
				msgBodyStr = body.Body
			} else {
				return fmt.Sprintf("解码消息内容异常，消息类型：%s", m.Msg.OperateDataType)
			}
			break
		}
	default:
		{
			msgBodyStr, err = json.Marshal(m.Msg.OperateData)
			if err != nil {
				return fmt.Sprintf("解码消息内容异常：%s", err.Error())
			}
			break
		}
	}
	m.endTimer = time.Now()
	subTimer := time.Since(m.inTimer)
	tps := calculateThroughput(1, subTimer.Milliseconds())
	result = fmt.Sprintf(
		"上下文标记：%s,创建标志：%s,输入时间：%s,截至时间：%s,耗时：%s ,处理速度： %.2f tps(事件数/秒),消息Base：%s,消息Body:%s",
		ctx,
		m.Flag,
		m.inTimer,
		m.endTimer,
		subTimer,
		tps,
		string(msgBaseStr),
		string(msgBodyStr),
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
