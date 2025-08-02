package mcp2eino

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

type ToolActionInfo struct {
	Id        string      `json:"id"`        // 运行id
	Name      string      `json:"name"`      // 工具名称
	Des       string      `json:"des"`       // 工具描述
	Args      string      `json:"args"`      // 工具参数
	Result    interface{} `json:"result"`    // 工具执行返回数据
	Message   string      `json:"message"`   // 工具执行结果(语音播报内容)
	Status    int         `json:"status"`    // 工具执行状态,0:正在执行，1:成功,2:失败
	TimeStamp int64       `json:"timeStamp"` // 时间戳
}
type EinoMcpMidInfo struct {
	SessionId      string              `json:"sessionId"`
	Err            error               `json:"err"`
	Result         *mcp.CallToolResult `json:"result"`
	ToolActionInfo ToolActionInfo      `json:"toolActionInfo"`
}
type If interface {
	Before(ctx context.Context, info *EinoMcpMidInfo)
	After(ctx context.Context, info *EinoMcpMidInfo)
}
type Before func(ctx context.Context, info *EinoMcpMidInfo)
type After func(ctx context.Context, info *EinoMcpMidInfo)

func NewMid(f1 Before, f2 After, log *zap.SugaredLogger) If {
	if f1 == nil || f2 == nil {
		panic("func is nil")
	}
	return &MidFun{
		f1:  f1,
		f2:  f2,
		log: log,
	}
}

type MidFun struct {
	f1  Before
	f2  After
	log *zap.SugaredLogger
}

func (m *MidFun) Before(ctx context.Context, info *EinoMcpMidInfo) {
	defer func() {
		if err := recover(); err != nil {
			if m.log != nil {
				m.log.Errorf("调用MCP工具执行前回调方法，出现异常:%v", err)
			} else {
				fmt.Printf("调用MCP工具执行前回调方法，出现异常:%v\n", err)
			}
		}
	}()
	m.f1(ctx, info)
}

func (m *MidFun) After(ctx context.Context, info *EinoMcpMidInfo) {
	defer func() {
		if err := recover(); err != nil {
			if m.log != nil {
				m.log.Errorf("调用MCP工具执行后回调方法，出现异常:%v", err)
			} else {
				fmt.Printf("调用MCP工具执行后回调方法，出现异常:%v\n", err)
			}
		}
	}()
	m.f2(ctx, info)
}

var _ If = (*MidFun)(nil)
