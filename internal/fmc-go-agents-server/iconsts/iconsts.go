package iconsts

// SubTopicList 需要订阅的主题
var SubTopicList = []string{
	Topic_Mqtt_To_Font_Cb,
}

// Mqtt的接收、发送消息的主题器
const (
	Topic_Mqtt_To_Font_Ca = "Topic_Mqtt_To_Font_Ca" // 与AI前端通信，发送给前端的主题，服务端->客户端
	Topic_Mqtt_To_Font_Cb = "Topic_Mqtt_To_Font_Cb" // 与AI前端通信，接收自前端的主题，客户端->客户端
)

// 业务消息主题
const (
	Topic_Bus_Page_Redirection_Within_The_System_Tool_Ca = "Topic_Mqtt_Page_Redirection_Within_The_System_Tool_Ca"
	Topic_Bus_Page_Redirection_Within_The_System_Tool_Cb = "Topic_Mqtt_Page_Redirection_Within_The_System_Tool_Cb"
)
