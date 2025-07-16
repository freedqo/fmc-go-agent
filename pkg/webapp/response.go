package webapp

// Response 接口响应
type Response struct {
	//错误码值
	Code int `json:"code"`
	//错误消息
	Message string `json:"message"`
	//响应数据
	Data interface{} `json:"data,omitempty"`
}
