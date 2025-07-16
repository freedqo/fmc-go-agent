package sm3

import (
	"encoding/hex"
	"fmt"
	"github.com/tjfoc/gmsm/sm3"
	"github.com/urfave/cli/v2"
)

var Handler = &cli.Command{
	Name:  "sm3",
	Usage: "生成sm3加密字符串",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "p",
			Usage:    "待加密的字符串",
			Aliases:  []string{},
			Required: true,
		},
	},
	Action: psm3,
}

// psm3 函数用于对输入的字符串进行 SM3 加密
func psm3(c *cli.Context) error {
	input := c.String("p")
	if input == "" {
		return fmt.Errorf("[错误] 请提供要加密的字符串作为参数")
	}

	// 创建 SM3 哈希对象
	h := sm3.New()

	// 写入要加密的数据
	fmt.Println("[信息] 正在写入数据到 SM3 哈希对象...")
	_, err := h.Write([]byte(input))
	if err != nil {
		return fmt.Errorf("[错误] 写入数据到 SM3 哈希对象时出错: %w", err)
	}

	// 计算哈希值
	fmt.Println("[信息] 正在计算 SM3 哈希值...")
	hashed := h.Sum(nil)

	// 将哈希值转换为十六进制字符串
	encrypted := hex.EncodeToString(hashed)

	// 输出结果
	fmt.Println("+-----------------------------------+")
	fmt.Printf("| 原始字符串      | %s\n", input)
	fmt.Printf("| SM3 加密结果    | %s\n", encrypted)
	fmt.Println("+-----------------------------------+")

	return nil
}
