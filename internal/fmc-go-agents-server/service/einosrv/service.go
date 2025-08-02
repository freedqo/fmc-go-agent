package einosrv

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/mcpsrv"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/store/log"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faiagent"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faiagent/mem"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	"go.uber.org/zap"
)

// New 函数用于创建一个新的Service实例
func New(ctx context.Context, opt *config.Config, dal dal.If, mcp mcpsrv.If, mem mem.MemoryIf, vp fmsg.PublishFunc) If {
	// 创建一个新的Service实例
	s := &Service{
		opt: opt,
		dal: dal,
		mem: mem,
		mcp: mcp,
		log: log.SysLog(),
		vp:  vp,
	}
	dbi := fvectordb.NewDocDbIf(dal.Db().Gdb().Urtyg_ai_agent().Ai_knowledge_documents().Self().UpdateDocumentsStatus, dal.Db().Gdb().Urtyg_ai_agent().Ai_knowledge_chunks().Self().SaveChunksData)
	vdb, err := fvectordb.New(ctx, opt.UVector, dbi)
	if err != nil {
		return nil
	}
	s.vdb = vdb
	// 返回Service实例
	return s
}

type Service struct {
	opt *config.Config
	log *zap.SugaredLogger
	dal dal.If
	mcp mcpsrv.If
	mem mem.MemoryIf
	vdb fvectordb.If
	vp  fmsg.PublishFunc
}

// UAiAgent 定义Service结构体的UAiAgent方法，返回uaiagent.If类型，可能需要给每个会话peek toolNode里面堵塞的流，所以不能复用了
func (s *Service) UAiAgent(ctx context.Context, sessionId string) (i faiagent.If, err error) {
	ts, err := s.mcp.GetEinoTools(ctx)
	if err != nil {
		return nil, err
	}
	return faiagent.New(ctx, s.log, s.dal.Cm(), s.mem, sessionId, ts, s.vdb), nil
}

// VectorDb 意图识别向量数据库，动态输入
// This function returns the vectorDb field of the Service struct
func (s *Service) VectorDb() fvectordb.If {
	return s.vdb
}
