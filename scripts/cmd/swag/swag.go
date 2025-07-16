package swag

import (
	"github.com/urfave/cli/v2"
)

var Handler = &cli.Command{
	Name:  "swag",
	Usage: "接口文档构建",
	Subcommands: []*cli.Command{
		{
			Name:   "fmt",
			Usage:  "整理注解格式",
			Action: FmtSwag,
		},
		{
			Name:  "gen",
			Usage: "根据注解生成接口文档",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "cmd",
					Aliases:  []string{"c"},
					Usage:    "指定cmd目录下的编译目录",
					Value:    "display-server-app",
					Required: true,
				},
			},
			Action: GenSwag,
		},
	},
}
