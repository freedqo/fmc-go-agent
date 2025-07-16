package cmdclear

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-cli/store/log"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-cli/store/uaivectordbx"
)

var Handler = &cli.Command{
	Name:    "clear",
	Usage:   "构建知识库",
	Aliases: []string{"c"},
	Flags:   []cli.Flag{},
	Action:  clearHandler,
}

func clearHandler(c *cli.Context) error {
	db := uaivectordbx.UAiVectorDb()
	log.SysLog().Infof("开始清理知识库")
	err := db.Redis().FlushDB(c.Context).Err()
	if err != nil {
		log.SysLog().Infof("清理知识库异常：%s", err.Error())
		return err
	}
	log.SysLog().Infof("清理知识库结束")

	return nil
}
