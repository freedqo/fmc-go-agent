package utils

import (
	"encoding/hex"
	"fmt"
	"github.com/tjfoc/gmsm/sm3"
)

func Sm3(input string) (string, error) {
	// 创建 SM3 哈希对象
	h := sm3.New()
	// 写入要加密的数据
	_, err := h.Write([]byte(input))
	if err != nil {
		return "", fmt.Errorf("写入数据到SM3哈希对象时出错: %w", err)
	}
	// 计算哈希值
	hashed := h.Sum(nil)
	// 将哈希值转换为十六进制字符串
	encrypted := hex.EncodeToString(hashed)
	return encrypted, nil
}
