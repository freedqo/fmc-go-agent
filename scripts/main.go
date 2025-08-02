package main

import (
	"fmt"
	"github.com/freedqo/fmc-go-agents/scripts/cmd/build"
	"github.com/freedqo/fmc-go-agents/scripts/cmd/dbgen"
	"github.com/freedqo/fmc-go-agents/scripts/cmd/receiverifgen"
	"github.com/freedqo/fmc-go-agents/scripts/cmd/sm3"
	"github.com/freedqo/fmc-go-agents/scripts/cmd/swag"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "FMC-GO-AGENT Code CLI",
		Usage: "FMC-GO-AGENT 辅助编程命令行工具",
		Commands: []*cli.Command{
			swag.Handler,          // 注册swag命令
			dbgen.Handler,         // 注册gen命令
			receiverifgen.Handler, // 注册receiverifgen命令
			build.Handler,         // 注册build命令
			sm3.Handler,           // 注册sm3命令
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
