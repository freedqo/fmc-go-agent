package dbgen

import (
	"github.com/freedqo/fmc-go-agents/scripts/cmd/dbgen/gencurd"
	"github.com/freedqo/fmc-go-agents/scripts/cmd/dbgen/genmodel"
	"github.com/urfave/cli/v2"
)

var Handler = &cli.Command{
	Name:  "gen",
	Usage: "数据库相关代码生成（模型+读写接口,不包含Web控制器）,需要使用大小写敏感的Mysql数据库",
	Subcommands: []*cli.Command{
		{
			Name:  "model",
			Usage: "生成模型文件(*.gen.go强制覆盖),需要使用大小写敏感的Mysql数据库",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "addr",
					Usage:    "指定Mysql数据库地址",
					Aliases:  []string{},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "user",
					Usage:    "指定用户",
					Aliases:  []string{},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "pwd",
					Usage:    "指定密码",
					Aliases:  []string{},
					Required: true,
				},
			},
			Action: genmodel.GenModel,
		},
		{
			Name:  "curd",
			Usage: "生成CURD接口(*.gen.go强制覆盖),需要使用大小写敏感的Mysql数据库",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "addr",
					Usage:    "指定Mysql数据库地址",
					Aliases:  []string{},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "user",
					Usage:    "指定用户",
					Aliases:  []string{},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "pwd",
					Usage:    "指定密码",
					Aliases:  []string{},
					Required: true,
				},
			},
			Action: gencurd.GenCURD,
		},
		{
			Name:  "all",
			Usage: "生成Model模型文件和CURD接口(*.gen.go强制覆盖),需要使用大小写敏感的Mysql数据库",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "addr",
					Usage:    "指定Mysql数据库地址",
					Aliases:  []string{},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "user",
					Usage:    "指定用户",
					Aliases:  []string{},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "pwd",
					Usage:    "指定密码",
					Aliases:  []string{},
					Required: true,
				},
			},
			Action: GenALL,
		},
	},
}

func GenALL(c *cli.Context) error {
	if err := genmodel.GenModel(c); err != nil {
		return err
	}
	if err := gencurd.GenCURD(c); err != nil {
		return err
	}
	return nil
}
