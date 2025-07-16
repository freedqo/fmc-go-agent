package swag

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"os/exec"
	"path/filepath"
)

func GenSwag(c *cli.Context) error {
	cmdDir := c.String("cmd")
	fmt.Println("[信息] 开始生成文档...")
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

	// 使用 filepath.Join 拼接路径
	mainGoPath := filepath.Join("cmd", cmdDir, "main.go")
	outputDocsPath := filepath.Join("cmd", cmdDir, "docs")

	// 打印路径以确认正确性
	fmt.Printf("[信息] 主 Go 文件路径: %s\n", mainGoPath)
	fmt.Printf("[信息] 文档输出目录: %s\n", outputDocsPath)

	// 构建命令及其参数
	cmdInit := exec.Command("swag", "init", "-g", mainGoPath, "-o", outputDocsPath)

	// 设置命令的工作目录为项目根目录
	cmdInit.Dir = rootDir

	fmt.Printf("[信息] 即将执行命令: %s\n", cmdInit.String())

	// 运行命令并捕获输出
	output, err := cmdInit.CombinedOutput()
	if err != nil {
		fmt.Printf("[错误] 命令执行失败: %v\n", err)
		fmt.Printf("[错误详情] %s\n", string(output))
		return fmt.Errorf("swag init 命令执行失败: %w", err)
	}

	// 打印命令输出结果
	fmt.Printf("[成功] 命令执行完毕\n%s\n", string(output))
	return nil
}

// findGoProjectRootDir 通过查找go.mod文件来确定项目的根目录
func findGoProjectRootDir(start string) (string, error) {
	current := start
	for {
		// 检查当前目录下是否存在go.mod文件
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}

		// 获取上一级目录
		next := filepath.Dir(current)

		// 如果已经到达根目录（即没有上一级目录），则返回错误
		if next == current {
			return "", fmt.Errorf("在任何父目录中都找不到 go.mod")
		}

		current = next
	}
}
