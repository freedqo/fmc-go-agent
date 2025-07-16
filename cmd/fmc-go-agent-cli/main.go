package main

import (
	"context"
	"fmt"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-cli/command/cmdbuild"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-cli/command/cmdclear"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-cli/store/log"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-cli/store/uaivectordbx"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/config"
	"github.com/freedqo/fmc-go-agent/pkg/uconfvp"
	"github.com/urfave/cli/v2"

	"os"
)

func main() {
	opt := config.NewDefault()
	_, err := uconfvp.New(opt, config.AppName)
	if err != nil {
		panic(err)
	}
	opt.UiRv.LoadMdFilePloy.IsLoadMdFiles = false
	ctx := context.Background()
	uaivectordbx.NewUAiVectorDb(ctx, opt.UiRv, log.SysLog())
	app := &cli.App{
		Name:  "fmc-go-agent-cli",
		Usage: "fmc-go-agent 命令行工具",
		Commands: []*cli.Command{
			cmdbuild.Handler,
			cmdclear.Handler,
		},
	}
	err = app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
