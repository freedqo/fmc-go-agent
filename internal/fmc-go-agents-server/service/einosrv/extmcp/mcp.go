package extmcp

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/iconsts"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/msgm"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/store/log"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp/faimcpclient"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp/faimcpclient/mcp2eino"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	"strings"
	"time"
)

type Object struct {
	Object interface{} `json:"object"`
}
type ToolActionInfo struct {
	Id        string      `json:"id"`        // 执行id，执行前后一致
	Name      string      `json:"name"`      // 工具名称
	Des       string      `json:"des"`       // 工具描述
	Args      string      `json:"args"`      // 工具参数
	Result    interface{} `json:"result"`    // 工具执行返回数据
	Message   string      `json:"message"`   // 工具执行结果(语音播报内容)
	Status    int         `json:"status"`    // 工具执行状态,0:正在执行，1:成功,2:失败
	TimeStamp int64       `json:"timeStamp"` // 时间戳
}

func New(ctx context.Context, mapMcp map[string]*faimcp.Option, vp fmsg.PublishFunc) If {
	mcpClient := make(map[string]faimcpclient.If)
	midFunc := mcp2eino.NewMid(
		func(ctx context.Context, info *mcp2eino.EinoMcpMidInfo) {
			if vp == nil {
				return
			}
			info.ToolActionInfo.Status = 0
			info.ToolActionInfo.TimeStamp = time.Now().Unix()
			msg := msgm.TAiAgentMessage{
				SessionId:   info.SessionId,
				MessageType: msgm.MessageType_Tool_Action,
				Topic:       "Topic_Tool_Action",
				RespType:    msgm.RespType_None,
				OperateID:   "",
				Data: Object{
					Object: info.ToolActionInfo,
				},
			}
			uMsg := fmsg.NewUMsg(iconsts.Topic_Mqtt_To_Font_Ca+"/"+info.SessionId, &msg, "McpServer", []string{fmsg.ToMqtt}, nil)
			vp(uMsg)

		},
		func(ctx context.Context, info *mcp2eino.EinoMcpMidInfo) {
			// 消息处理器没注入，不处理
			if vp == nil {
				return
			}
			// 处理调用的异常
			if info.Err != nil || info.Result == nil {
				info.ToolActionInfo.TimeStamp = time.Now().Unix()
				info.ToolActionInfo.Status = 2
				if info.Err != nil {
					info.ToolActionInfo.Result = info.Err
				}
				log.SysLog().Errorf("mcp客户端，调用工具%s,异常:%v", info.ToolActionInfo.Name, info)
				info.ToolActionInfo.Message = fmt.Sprintf("执行报错,请联系管理员处理")
				//执行语音汇报功能
				msg := msgm.TAiAgentMessage{
					SessionId:   info.SessionId,
					MessageType: msgm.MessageType_Tool_Action,
					Topic:       "Topic_Tool_Action",
					RespType:    msgm.RespType_None,
					OperateID:   "",
					Data: Object{
						Object: info.ToolActionInfo,
					},
				}
				uMsg := fmsg.NewUMsg(iconsts.Topic_Mqtt_To_Font_Ca+"/"+info.SessionId, &msg, "McpServer", []string{fmsg.ToMqtt}, nil)
				vp(uMsg)
			}
			// 处理约定好的应答结果
			defer func() {
				info.ToolActionInfo.TimeStamp = time.Now().Unix()
				//执行语音汇报功能
				msg := msgm.TAiAgentMessage{
					SessionId:   info.SessionId,
					MessageType: msgm.MessageType_Tool_Action,
					Topic:       "Topic_Tool_Action",
					RespType:    msgm.RespType_None,
					OperateID:   "",
					Data: Object{
						Object: info.ToolActionInfo,
					},
				}
				uMsg := fmsg.NewUMsg(iconsts.Topic_Mqtt_To_Font_Ca+"/"+info.SessionId, &msg, "McpServer", []string{fmsg.ToMqtt}, nil)
				vp(uMsg)
			}()
			// 处理约定好的应答结果
			if info.Result.Meta == nil {
				log.SysLog().Errorf("mcp客户端，调用工具%s,异常:工具链路未注入_meta", info.ToolActionInfo.Name)
				info.ToolActionInfo.Status = 2
				info.ToolActionInfo.Result = info.Result
				info.ToolActionInfo.Message = fmt.Sprintf("执行异常,未注入应答meta")
				return
			}
			sessionId, ok := info.Result.Meta["sessionId"].(string)
			if !ok || strings.TrimSpace(sessionId) == "" || sessionId != info.SessionId {
				log.SysLog().Errorf("mcp客户端，调用工具%s执行前中间件,异常:工具链路meta.sessionId异常,%v", info.ToolActionInfo.Name, info.Result.Meta)
				info.ToolActionInfo.Status = 2
				info.ToolActionInfo.Result = info.Result
				info.ToolActionInfo.Message = fmt.Sprintf("执行异常,应答meta.sessionId异常")
				return
			}
			status, ok := info.Result.Meta["status"].(float64)
			if !ok {
				log.SysLog().Errorf("mcp客户端，调用工具%s,异常:工具链路meta.status异常,%v", info.ToolActionInfo.Name, info.Result.Meta)
				info.ToolActionInfo.Status = 2
				info.ToolActionInfo.Result = info.Result
				info.ToolActionInfo.Message = fmt.Sprintf("执行异常,应答meta.status异常")
				return
			}
			info.ToolActionInfo.Status = int(status)

			message, ok := info.Result.Meta["message"].(string)
			if !ok || strings.TrimSpace(message) == "" {
				log.SysLog().Errorf("mcp客户端，调用工具%s,异常:工具链路meta.message异常,%v", info.ToolActionInfo.Name, info.Result.Meta)
				info.ToolActionInfo.Status = 2
				info.ToolActionInfo.Result = info.Result
				info.ToolActionInfo.Message = fmt.Sprintf("执行异常,应答meta.message异常")
				return
			}
			info.ToolActionInfo.Message = message
			info.ToolActionInfo.Result = info.Result
			return
		},
		log.SysLog(),
	)

	for k, v := range mapMcp {
		mc, err := faimcpclient.New(ctx, k, v, log.SysLog())
		if err != nil {
			log.SysLog().Errorf("mcp客户端，初始化异常:%s", err.Error())
			continue
		}
		mcpClient[k] = mc
		mcpClient[k].SubToolMidFunc(&midFunc)
	}
	return &ExtMCP{
		mcp: mcpClient,
		vp:  vp,
	}
}

type ExtMCP struct {
	mcp map[string]faimcpclient.If
	vp  fmsg.PublishFunc
}

func (e *ExtMCP) EinoTools(ctx context.Context) []tool.BaseTool {
	tools := make([]tool.BaseTool, 0)
	for _, v := range e.mcp {
		var v1 faimcpclient.If
		v1 = v
		tools = append(tools, v1.DToEinoTools(ctx)...)
	}
	return tools
}

var _ If = &ExtMCP{}
