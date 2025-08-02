// message
package fmsg

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
)

// --------------------------------消息基本信息---------------
type MessageBase struct {
	ClientID        string `json:"clientID"`        //客户端ID(guid)
	Operate         string `json:"operate"`         //操作（空值用于表示握手）。如果是应答消息，Operate可以和接收的消息的Operate一致。
	Retained        bool   `json:"retained"`        //是否保留,mqtt协议中保留消息
	IsReplyOperate  bool   `json:"isReplyOperate"`  //是否应答操作
	OperateID       string `json:"operateID"`       //操作ID(guid)。如果是应答消息，OperateID应和接收的消息的OperateID保持一致。
	OperateDataType string `json:"operateDataType"` //操作数据类型。如果是应答消息，则指应答数据内容类型。一般情况下，OperateDataType和接收的消息的OperateDataType是不一样的。
}

func (m MessageBase) String() string {
	return fmt.Sprintf("{ClientID=%s, Operate=%s, IsReplyOperate=%v, OperateID=%s, OperateDataType=%s}", m.ClientID, m.Operate, m.IsReplyOperate, m.OperateID, m.OperateDataType)
}

// --------------------------------------------消息-----------------------------------
// 构建消息实例，并自动分配OperateID。
// 该消息为通用情况：用于发送、应答、json解码等
func NewMessage() Message {
	var msg Message
	msg.OperateID = uuid.New().String()
	return msg
}

// 构建用于发送的消息实例。
// 自动分配OperateID，并将IsReplyOperate设置为false。
func NewMessageForSend(clientID, operate, operateDataType string, operateData interface{}) Message {
	var msg Message
	msg.ClientID = clientID
	msg.Operate = operate
	msg.OperateID = uuid.New().String()
	msg.IsReplyOperate = false
	msg.OperateDataType = operateDataType
	msg.OperateData = operateData
	return msg
}

// 构建用于应答的消息实例。
// 将IsReplyOperate设置为true。
func NewMessageForReply(clientID, operate, operateID, operateDataType string, operateData interface{}) Message {
	var msg Message
	msg.ClientID = clientID
	msg.Operate = operate
	msg.IsReplyOperate = true
	msg.OperateID = operateID
	msg.OperateDataType = operateDataType
	msg.OperateData = operateData
	return msg
}

// 透明传输时，打包通用Message消息的数据类型
const TransparentTransport_package_message_OperateDataType = "[]byte"

/*
将msgData打包成通用Message消息（即OperateDataType = "[]byte"）
说明，如果对方（websocket服务端）的消息不采用Message消息格式，则采用这种机制进行打包，以实现透明传输。
*/
func NewMessageForPackage(msgData []byte) Message {
	var msg = NewMessage()
	msg.OperateDataType = TransparentTransport_package_message_OperateDataType
	msg.OperateData = msgData
	return msg
}

// 消息
type Message struct {
	MessageBase
	OperateData interface{} `json:"operateData"` //操作数据指针
}

func (m Message) String() string {
	return fmt.Sprintf("{ClientID=%s, Operate=%s, IsReplyOperate=%v, OperateID=%s, OperateDataType=%s,OperateData=%v}", m.ClientID, m.Operate, m.IsReplyOperate, m.OperateID, m.OperateDataType, m.OperateData)
}

//--------------------------------------------输入消息-----------------------------------

// 构建输入消息
func NewInMessage(msgBase MessageBase, messageData interface{}) (imsg InMessage, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()
	var inMsg InMessage
	inMsg.BaseInfo = msgBase
	inMsg.MessageData = messageData
	return inMsg, nil
}

// 输入消息
type InMessage struct {
	BaseInfo    MessageBase //从MessageData解析的消息的基本信息
	MessageData interface{} //Message结构体的json编码的数据流。可以用于解码：在解码前，请根据BaseInfo.OperateDataType,给抽象变量：Message.OperateData赋值具体变量指针，然后再解码
}

func (m InMessage) String() string {
	return fmt.Sprintf("{BaseInfo=%v, MessageData =%v}", m.BaseInfo, m.MessageData)
}

//--------------------------------------------输出消息-----------------------------------

// 构建输出消息
func NewOutMessage(msg Message) (omsg OutMessage, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()
	data, err := json.Marshal(msg)
	if err != nil {
		return omsg, err
	}
	var outMsg OutMessage
	outMsg.ClientID = msg.ClientID
	outMsg.MessageData = make([]byte, 0, len(data))
	outMsg.MessageData = append(outMsg.MessageData, data...)
	return outMsg, nil
}

// 输出消息
type OutMessage struct {
	ClientID    string //客户端ID。如果为空，表示广播发送消息
	MessageData []byte //Message结构体的json编码的数据流。
}

func (m OutMessage) String() string {
	return fmt.Sprintf("{ClientID=%s, MessageData size=%d}", m.ClientID, len(m.MessageData))
}

// 创建客户端连接状态
func NewClientConnectState(clientID string, isConnected bool, remoteAddr string) ClientConnectState {
	return ClientConnectState{
		ClientID:    clientID,
		IsConnected: isConnected,
		RemoteAddr:  remoteAddr,
	}
}

// ClientConnectState 客户端连接状态
type ClientConnectState struct {
	ClientID    string //客户端ID
	IsConnected bool   //是否连接
	RemoteAddr  string //
}
