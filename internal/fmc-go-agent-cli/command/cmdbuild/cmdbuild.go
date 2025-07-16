package cmdbuild

import (
	"fmt"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-cli/store/log"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-cli/store/uaivectordbx"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

var Handler = &cli.Command{
	Name:    "build",
	Usage:   "构建知识库",
	Aliases: []string{"b"},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "dir",
			Aliases:  []string{"d"},
			Usage:    "指定需要构建的文档路径",
			Value:    "./knowdb/md",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "select",
			Aliases:  []string{"s"},
			Usage:    "选择指定构建文件",
			Required: false,
		},
		&cli.BoolFlag{
			Name:    "isClear",
			Aliases: []string{"ic"},
			Usage:   "构建前是否清理向量数据库缓存的数据",
			Value:   false,
		},
	},
	Action: buildHandler,
}

func buildHandler(c *cli.Context) error {
	dir := c.String("dir")
	if dir == "" {
		return fmt.Errorf("dir is required")
	}
	file := c.String("select")
	isClear := c.Bool("isClear")

	db := uaivectordbx.UAiVectorDb()
	if isClear {
		err := db.Redis().FlushDB(c.Context).Err()
		if err != nil {
			return err
		}
	}
	if file != "" {
		//TODO:构建指定文件
		err := db.BuildFile(c.Context, path.Join(dir, file))
		if err != nil {
			return err
		}
	} else {
		log.SysLog().Infof("开始构建目录 %s\n", dir)
		// 打印文件夹下面所有的文件名称
		// 检查目录是否存在
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.SysLog().Infof("错误：目录 %s 不存在\n", dir)
			return err
		}
		// 遍历目录并列出文件
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// 如果不是递归模式且当前项是目录，则跳过其内容
			if d.IsDir() {
				// 跳过子目录的内容
				return nil
			}

			// 获取相对于指定目录的相对路径
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			// 打印文件名（或目录名）
			if d.IsDir() {
				log.SysLog().Infof("目录: %s/\n", relPath)
			} else {
				log.SysLog().Infof("文件: %s\n", relPath)
			}

			return nil
		})
		//TODO:构建全部
		err = db.BuildDir(c.Context, dir)
		if err != nil {
			return err
		}
	}
	return nil
}
