package cmdbuild

import (
	"fmt"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"sync"

	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-cli/store/log"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-cli/store/urag"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/dbm/urtyg_ai_agent/model"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb"
	"github.com/urfave/cli/v2"
	"io/fs"
	"os"
	"path/filepath"
	"time"
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
	Action: build,
}

func build(c *cli.Context) error {
	rag, db := urag.GetURag(c.Context)
	dir := c.String("dir")
	if dir == "" {
		return fmt.Errorf("dir is required")
	}
	//file := c.String("select")
	//isClear := c.Bool("isClear")
	log.SysLog().Infof("开始构建目录 %s\n", dir)
	// 打印文件夹下面所有的文件名称
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.SysLog().Infof("错误：目录 %s 不存在\n", dir)
		return err
	}
	wg := sync.WaitGroup{}
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
			return nil
		} else {
			log.SysLog().Infof("开始构建文件: %s\n", relPath)
		}
		fileInfo, err := d.Info()
		if err != nil {
			return err
		}
		knowledgeBaseName := "odq"
		now := time.Now()
		id := int64(utils.GetIntID())
		err = db.Gdb().Urtyg_ai_agent().Ai_knowledge_documents().Gen().Save(&model.Ai_knowledge_documents{
			ID:                id,
			KnowledgeBaseName: knowledgeBaseName,
			FileName:          fileInfo.Name(),
			Status:            0,
			CreatedAt:         &now,
			UpdatedAt:         &now,
		})
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			ids, err := rag.Index(c.Context, &fvectordb.IndexReq{
				URI:           path,
				KnowledgeName: knowledgeBaseName,
				DocumentsId:   id,
			})
			if err != nil {
				fmt.Printf("构建异常：%v", err)
				return
			}
			fmt.Printf("构建成功,id列表：%s\n", ids)
		}()

		return nil
	})
	wg.Wait()
	if err != nil {
		return err
	}
	return nil
}
