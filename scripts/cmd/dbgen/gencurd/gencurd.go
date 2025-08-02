package gencurd

import (
	"github.com/freedqo/fmc-go-agents/scripts/cmd/dbgen/gdb"
	"github.com/freedqo/fmc-go-agents/scripts/cmd/dbgen/options"
	"github.com/freedqo/fmc-go-agents/scripts/cmd/dbgen/utils"
	"github.com/urfave/cli/v2"
)

func GenCURD(c *cli.Context) error {
	for _, db := range options.DbList {
		err := CreatCurdCode(c.String("addr"), db, c.String("user"), c.String("pwd"))
		if err != nil {
			return err
		}
	}
	DbMap := make(map[string]string, 0)
	for _, db := range options.DbList {
		DbMap[db] = utils.FirstToUpper(db)
	}
	err := CreatAllDbIf(options.DbIfRootPath, DbMap)
	if err != nil {
		return err
	}
	err = CreatGdbAllDbImp(options.DbGdbRootPath, DbMap)
	if err != nil {
		return err
	}

	return nil
}
func CreatCurdCode(addr string, dbname string, user string, password string) error {
	_, _, _, dbMeta, err := gdb.NewGdbConn(addr, dbname, user, password)
	if err != nil {
		return err
	}
	err = CreateIfFile(dbname, dbMeta)
	if err != nil {
		return err
	}
	err = CreatGdbDbImp(dbname, dbMeta)
	if err != nil {
		return err
	}
	return nil
}
