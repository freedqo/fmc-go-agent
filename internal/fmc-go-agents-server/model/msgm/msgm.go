package msgm

type MessageType int

const (
	MessageType_Chat             MessageType = iota // 聊天
	MessageType_Page_Redirection                    // 页面跳转
	MessageType_Bus_Operate                         // 业务操作
	MessageType_Voice_Broadcast                     // 语音播报
	MessageType_Tool_Action                         // 工具执行情况
)

type RespType int

const (
	RespType_None             RespType = iota // 无需应答
	RespType_ServerReqArk                     //1-2:服务端发送应答流,服务端发送1，
	RespType_ClientRespServer                 //1-2:接收端发送2,应答服务端
	RespType_ClientReqArk                     //3-4:客户端发送应答流,客户端发送3
	RespType_ServerRespClient                 //3-4:需要服务端发送4,应答客户端
)

type TAiAgentMessage struct {
	SessionId   string      `json:"sessionId"`   //会话ID
	MessageType MessageType `json:"messageType"` //业务类型
	Topic       string      `json:"topic"`       //消息主题
	//应答类型:
	//0,单向消息流,接收端不需要应答;
	//1:服务端发送应答流,服务端发送1，需要接收端发送2,应答服务端
	//3:客户端发送应答流,客户端发送3，需要服务端发送4,应答客户端
	RespType  RespType    `json:"respType"`
	OperateID string      `json:"operateID"` //操作ID。如果是应答消息，OperateID应和接收的消息的OperateID保持一致。
	Data      interface{} `json:"data"`      //操作数据指针
}
