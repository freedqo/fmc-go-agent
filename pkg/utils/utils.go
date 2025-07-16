package utils

import (
	"github.com/gorilla/websocket"
	"net/url"
)

// CheckWebSocketService 检测 WebSocket 服务是否可用
func CheckWebSocketService(serviceURL string) error {
	u, err := url.Parse(serviceURL)
	if err != nil {
		return err
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	err = conn.Close()
	return err
}

func SplitString(s string, maxLength int) []string {
	var result []string
	for len(s) > 0 {
		if len(s) <= maxLength {
			result = append(result, s)
			break
		}
		result = append(result, s[:maxLength])
		s = s[maxLength:]
	}
	return result
}

