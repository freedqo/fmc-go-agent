package main

import (
	"fmt"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-cli/command/cmdRetriever"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-cli/command/cmdbuild"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-cli/command/cmdclear"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	"github.com/freedqo/fmc-go-agents/pkg/fconfvp"

	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	opt := config.NewDefault()
	_, err := fconfvp.New(opt, config.AppName)
	if err != nil {
		panic(err)
	}
	app := &cli.App{
		Name:  "fmc-go-agents-cli",
		Usage: "知识库构建命令行工具",
		Commands: []*cli.Command{
			cmdbuild.Handler,
			cmdclear.Handler,
			cmdRetriever.Handler,
		},
	}
	err = app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
