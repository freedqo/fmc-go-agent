package swag

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"os/exec"
)

func FmtSwag(c *cli.Context) error {
	// 获取当前工作目录
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("错误：无法获取当前工作目录: %w", err)
	}
	fmt.Println("[信息] 已获取当前工作目录:", pwd)

	// 计算项目根目录的路径
	rootDir, err := findGoProjectRootDir(pwd)
	if err != nil {
		return fmt.Errorf("错误：无法确定项目根目录: %w", err)
	}
	fmt.Println("[信息] 已确定项目根目录:", rootDir)

	// 构建命令及其参数
	cmdInit := exec.Command("swag", "fmt")

	// 设置命令的工作目录为项目根目录
	cmdInit.Dir = rootDir

	fmt.Printf("[信息] 即将执行命令: %s\n", cmdInit.String())

	// 运行命令并捕获输出
	output, err := cmdInit.CombinedOutput()
	if err != nil {
		fmt.Printf("[错误] 命令执行失败: %v\n", err)
		fmt.Printf("[错误详情] %s\n", string(output))
		return fmt.Errorf("swag fmt 命令执行失败: %w", err)
	}

	// 打印命令输出结果
	fmt.Printf("[成功] 命令执行完毕\n%s\n", string(output))
	return nil
}
