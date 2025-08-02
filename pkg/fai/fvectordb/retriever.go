package fvectordb

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb/common"
	"sort"
	"sync"

	"github.com/cloudwego/eino-ext/components/retriever/es8"
	er "github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

type RetrieveReq struct {
	Query         string   // 检索关键词
	TopK          int      // 检索结果数量
	Score         float64  //  分数阀值(0-2, 0 完全相反，1 毫不相干，2 完全相同,一般需要传入一个大于1的数字，如1.5)
	KnowledgeName string   // 知识库名字
	optQuery      string   // 优化后的检索关键词
	excludeIDs    []string // 要排除的 _id 列表
	rankScore     float64  // 排名分数，原本的score是0-2（实际是1-2），需要在这里改成0-1
}

func (x *RetrieveReq) copy() *RetrieveReq {
	return &RetrieveReq{
		Query:         x.Query,
		TopK:          x.TopK,
		Score:         x.Score,
		KnowledgeName: x.KnowledgeName,
		optQuery:      x.optQuery,
		excludeIDs:    x.excludeIDs,
		rankScore:     x.rankScore,
	}
}

// Retrieve 检索
func (x *FRag) Retrieve(ctx context.Context, query string, opts ...er.Option) ([]*schema.Document, error) {
	var (
		used        = ""          // 记录已经使用过的关键词
		relatedDocs = &sync.Map{} // 记录相关docs
	)
	msg := make([]*schema.Document, 0)
	req := &RetrieveReq{
		Query:         query,
		TopK:          5,
		Score:         1.2,
		KnowledgeName: "odq",
	}
	req.rankScore = req.Score
	// 大于1的需要-1
	if req.rankScore >= 1 {
		req.rankScore -= 1
	}
	rewriteModel, err := common.GetRewriteModel(ctx, &openai.ChatModelConfig{
		Model:   x.opt.Rewrite.Model,
		BaseURL: x.opt.Rewrite.BaseURL,
		APIKey:  x.opt.Rewrite.APIKey,
	})
	if err != nil {
		return nil, err
	}
	wg := &sync.WaitGroup{}
	// 尝试N次重写关键词进行搜索,后续可以考虑做成配置
	for i := 0; i < 3; i++ {
		question := req.Query
		var (
			optMessages    []*schema.Message
			rewriteMessage *schema.Message
		)
		optMessages, err = getOptimizedQueryMessages(used, question, req.KnowledgeName)
		if err != nil {
			return nil, err
		}
		rewriteMessage, err = rewriteModel.Generate(ctx, optMessages)
		if err != nil {
			return nil, err
		}
		optimizedQuery := rewriteMessage.Content
		used += optimizedQuery + " "
		req.optQuery = req.Query
		wg.Add(1)
		go func() {
			defer wg.Done()
			rDocs := make([]*schema.Document, 0)
			rDocs, err = x.retrieveDoOnce(ctx, req.copy())
			if err != nil {
				return
			}
			for _, doc := range rDocs {
				if old, e := relatedDocs.LoadOrStore(doc.ID, doc); e {
					// 同文档则保存较高分的结果（对于不同的optQuery，rerank可能会有不同的结果）
					if doc.Score() > old.(*schema.Document).Score() {
						relatedDocs.Store(doc.ID, doc)
					}
				}
			}

		}()
	}
	wg.Wait()
	// 整理需要返回的数据
	relatedDocs.Range(func(key, value any) bool {
		msg = append(msg, value.(*schema.Document))
		return true
	})
	sort.Slice(msg, func(i, j int) bool {
		return msg[i].Score() > msg[j].Score()
	})
	if len(msg) > req.TopK {
		msg = msg[:req.TopK]
	}
	return msg, nil
}

func (x *FRag) retrieveDoOnce(ctx context.Context, req *RetrieveReq) (relatedDocs []*schema.Document, err error) {
	var (
		docs   []*schema.Document
		qaDocs []*schema.Document
	)
	// 通过内容检索
	docs, err = x.retrieve(ctx, req, false)
	if err != nil {
		return
	}
	if x.opt.Rewrite != nil {
		// 通过qa检索
		qaDocs, err = x.retrieve(ctx, req, true)
		if err != nil {
			return
		}
		docs = append(docs, qaDocs...)
	}

	// 去重
	docs = common.RemoveDuplicates(docs, func(doc *schema.Document) string {
		return doc.ID
	})
	//// 重排
	//docs, err = rerank.NewRerank(ctx, req.optQuery, docs, req.TopK)
	//if err != nil {
	//	return
	//}
	for _, doc := range docs {
		if doc.Score() < req.rankScore {
			continue
		}
		relatedDocs = append(relatedDocs, doc)
	}
	return
}

func (x *FRag) retrieve(ctx context.Context, req *RetrieveReq, qa bool) (msg []*schema.Document, err error) {
	esQuery := []types.Query{
		{
			Bool: &types.BoolQuery{
				Must: []types.Query{{Match: map[string]types.MatchQuery{common.KnowledgeName: {Query: req.KnowledgeName}}}},
			},
		},
	}
	if len(req.excludeIDs) > 0 {
		esQuery[0].Bool.MustNot = []types.Query{
			{
				Terms: &types.TermsQuery{
					TermsQuery: map[string]types.TermsQueryField{
						"_id": req.excludeIDs,
					},
				},
			},
		}
	}
	r := x.retriever
	if qa {
		r = x.qaRetriever
	}
	msg, err = r.Invoke(ctx, req.optQuery,
		compose.WithRetrieverOption(
			// er.WithScoreThreshold(req.Score), // 不限制分数，只限制数量，最终分数由rerank给
			er.WithTopK(esTopK),
			es8.WithFilters(esQuery),
		),
	)
	for _, s := range msg {
		if s.Score() > 1 {
			// 本身没意义，最终分数由rerank给，这里只是为了方便测试观察
			s.WithScore(s.Score() - 1)
		}
	}
	if err != nil {
		return
	}
	return
}
