package mcpsrv

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/iconsts"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/dbm/urtyg_ai_agent/model"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/msgm"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/store/log"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp/faimcpclient"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp/faimcpclient/mcp2eino"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"strings"
	"time"
)

type Object struct {
	Object interface{} `json:"object"`
}

func (s *MCPSrv) Before(ctx context.Context, info *mcp2eino.EinoMcpMidInfo) {
	if s.vp == nil {
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
	s.vp(uMsg)

}
func (s *MCPSrv) After(ctx context.Context, info *mcp2eino.EinoMcpMidInfo) {
	// 消息处理器没注入，不处理
	if s.vp == nil {
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
		s.vp(uMsg)
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
		s.vp(uMsg)
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
}

func (s *MCPSrv) newMCPClients(ctx context.Context) map[string]faimcpclient.If {
	if s.opt.MCP.Clients != nil {
		mcpCMap := make(map[string]faimcpclient.If)
		midFunc := mcp2eino.NewMid(
			s.Before,
			s.After,
			log.SysLog(),
		)
		for k, v := range s.opt.MCP.Clients {
			mc, err := faimcpclient.New(ctx, k, v, log.SysLog())
			if err != nil {
				log.SysLog().Errorf("mcp客户端，初始化异常:%s", err.Error())
				continue
			}
			mcpCMap[k] = mc
			mcpCMap[k].SubToolMidFunc(&midFunc)
		}
		return mcpCMap
	}
	return nil
}

func (s *MCPSrv) GetEinoTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := make([]tool.BaseTool, 0)
	clients := s.newMCPClients(ctx)
	if clients == nil {
		return tools, nil
	}
	// 查mcp列表
	mcpList, _, err := s.db.Gdb().Urtyg_ai_agent().Ai_mcp().Gen().Find(nil)
	if err != nil {
		return nil, err
	}
	// 查工具列表
	aiTools, _, err := s.db.Gdb().Urtyg_ai_agent().Ai_tool().Gen().Find(nil)
	if err != nil {
		return nil, err
	}
	for _, v := range clients {
		var v1 faimcpclient.If
		v1 = v
		ts := v1.DToEinoTools(ctx)
		ts1, err := s.SaveMcpTool(ctx, v, ts, mcpList, aiTools)
		if err != nil {
			return nil, err
		}
		tools = append(tools, ts1...)
	}
	return tools, nil
}
func (s *MCPSrv) SaveMcpTool(ctx context.Context, mcpc faimcpclient.If, ts []tool.BaseTool, mcpList []*model.Ai_mcp, aiTools []*model.Ai_tool) ([]tool.BaseTool, error) {
	tools := make([]tool.BaseTool, 0)
	mi := mcpc.ServerInfo()
	index := -1
	for i, m := range mcpList {
		if m.Name == mi.ServerInfo.Name {
			index = i
			break
		}
	}
	mcpid := ""
	if index != -1 {
		// 存在
		mcpid = mcpList[index].ID
		mcpList[index].Version = mi.ServerInfo.Version
		mcpList[index].UpdatedAt = time.Now()
		// 更新
		err := s.db.Gdb().Urtyg_ai_agent().Ai_mcp().Gen().Upt(mcpList[index])
		if err != nil {
			return nil, err
		}
	} else {
		// 不存在
		mcpid = utils.GetStringID()
		err := s.db.Gdb().Urtyg_ai_agent().Ai_mcp().Gen().Add(&model.Ai_mcp{
			ID:        mcpid,
			Type:      0,
			Enable:    true,
			Name:      mi.ServerInfo.Name,
			Version:   mi.ServerInfo.Version,
			Des:       mi.ServerInfo.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			return nil, err
		}
	}
	for _, t := range ts {
		ti, err := t.Info(ctx)
		if err != nil {
			return nil, err
		}
		// 检查工具是否已经存在
		isExit := false
		isEnable := false
		for _, aiTool := range aiTools {
			if aiTool.Name == ti.Name {
				isExit = true
				if aiTool.Enable {
					isEnable = true
				}
				break
			}
		}
		if !isExit {
			err = s.db.Gdb().Urtyg_ai_agent().Ai_tool().Gen().Add(&model.Ai_tool{
				ID:        utils.GetStringID(),
				Enable:    true,
				Name:      ti.Name,
				Des:       ti.Desc,
				McpID:     mcpid,
				CreatedAt: utils.Of(time.Now()),
				UpdatedAt: utils.Of(time.Now()),
			})
			if err != nil {
				return nil, err
			}
		}
		if isEnable {
			tools = append(tools, t)
		}
	}
	return tools, nil
}
