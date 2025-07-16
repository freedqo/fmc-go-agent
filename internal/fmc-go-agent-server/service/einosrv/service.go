package einosrv

import (
	"context"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/config"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/einosrv/extmcp"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/log"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaiagent"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaiagent/mem"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaivectordb"
	"go.uber.org/zap"
)

// New 函数用于创建一个新的Service实例
func New(ctx context.Context, opt *config.Config, dal dal.If, mem mem.MemoryIf) If {
	// 创建一个新的Service实例
	s := &Service{
		vectorDb: uaivectordb.New(ctx, opt.UiRv, log.SysLog()), // 初始化iRVector字段，使用uaiirvector.New函数创建一个新的uaiirvector实例
		dal:      dal,
		mem:      mem,
		log:      log.SysLog(),
	}
	// 添加工具集合
	extMCP := extmcp.New(ctx, opt.Ext.McpServer)
	// 检查extMCP是否为nil，如果是nil，则panic
	if extMCP == nil {
		panic("extMCP is nil")
	}
	s.extMcp = extMCP

	// 返回Service实例
	return s
}

type Service struct {
	log      *zap.SugaredLogger
	vectorDb uaivectordb.If //ai向量（数据库）知识库-热载对象
	dal      dal.If
	mem      mem.MemoryIf
	extMcp   extmcp.If
}

// UAiAgent 定义Service结构体的UAiAgent方法，返回uaiagent.If类型，可能需要给每个会话peek toolNode里面堵塞的流，所以不能复用了
func (s *Service) UAiAgent(ctx context.Context, sessionId string) uaiagent.If {
	// 使用uaiagent.New方法创建一个新的uaiagent实例，传入context、s.log、s.dal.Cm()、s.vectorDb.Embedder()、s.vectorDb.Redis()、s.mem、s.extMcp.EinoTools(ctx)作为参数
	return uaiagent.New(ctx, s.log, s.dal.Cm(), s.vectorDb.Embedder(), s.vectorDb.Redis(), s.mem, sessionId, s.extMcp.EinoTools(ctx))
}

// VectorDb 意图识别向量数据库，动态输入
func (s *Service) VectorDb() uaivectordb.If {
	return s.vectorDb
}
