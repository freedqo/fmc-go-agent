package cmdclear

import (
	"fmt"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-cli/store/urag"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/dbm/urtyg_ai_agent/model"
	"github.com/urfave/cli/v2"
	"strconv"
)

var Handler = &cli.Command{
	Name:    "clear",
	Usage:   "构建知识库",
	Aliases: []string{"c"},
	Flags:   []cli.Flag{},
	Action:  clear,
}

func clear(c *cli.Context) error {
	rag, db := urag.GetURag(c.Context)
	documents, _, err := db.Gdb().Urtyg_ai_agent().Ai_knowledge_documents().Gen().Find(nil)
	if err != nil {
		return err
	}
	fmt.Printf("开始清理知识库，共%d个文档\n", len(documents))
	for _, v := range documents {
		fmt.Printf("删除文档：%s\n", v.FileName)
		id := strconv.FormatInt(v.ID, 10)
		chunks, _, err := db.Gdb().Urtyg_ai_agent().Ai_knowledge_chunks().Gen().Find(&model.Ai_knowledge_chunks_QueryReq{
			KnowledgeDocID: &id,
		})
		if err != nil {
			return err
		}
		for i, v1 := range chunks {
			fmt.Printf("删除文档：%s,切片：%d\n", v.FileName, i)
			err = rag.DeleteDocument(c.Context, v1.ChunkID)
			if err != nil {
				return err
			}
		}
		err = db.Gdb().Urtyg_ai_agent().Ai_knowledge_documents().Gen().Del(&model.Ai_knowledge_documents{ID: v.ID})
		if err != nil {
			return err
		}

	}
	return nil
}
