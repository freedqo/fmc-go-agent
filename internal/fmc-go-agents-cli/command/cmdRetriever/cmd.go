package cmdRetriever

import (
	"bufio"
	"fmt"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-cli/store/log"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-cli/store/urag"
	"github.com/urfave/cli/v2"
	"os"
)

var Handler = &cli.Command{
	Name:    "queryKdb",
	Usage:   "查询知识库",
	Aliases: []string{"q"},
	Flags:   []cli.Flag{},
	Action:  query,
}

// 清理知识库
func query(c *cli.Context) error {
	// 获取URag和数据库
	rag, _ := urag.GetURag(c.Context)
	log.SysLog().Infof("开始检索知识库")
	sc := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("输入检索内容(exit退出)>")
		sc.Scan()
		if err := sc.Err(); err != nil {
			return err
		}
		input := sc.Text()
		if input == "exit" {
			break
		}
		documents, err := rag.Retrieve(c.Context, input)
		if err != nil {
			return err
		}
		if documents == nil || len(documents) == 0 {
			fmt.Printf("检索知识库未发现\n")
		} else {
			fmt.Printf("检索知识库发现%d条数据\n", len(documents))
			for i, document := range documents {
				fmt.Printf("检索知识库发现-[%d]->id:%s,Content:%s\n", i, document.ID, document.Content)
			}
		}
	}
	return nil
}
