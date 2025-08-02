package mcpsrv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/config"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/db"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/ext"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/iconsts"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/dalm/extm/Knowledgem"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/msgm"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/store/log"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/store/urecover"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp/faimcpclient"
	"github.com/freedqo/fmc-go-agents/pkg/fai/faimcp/faimcpserver"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"time"
)

func New(ctx context.Context, opt *config.Config, dal dal.If, vp fmsg.PublishFunc, vwr func(operateID string, waitTime time.Duration) (*fmsg.UMsg, error)) If {
	if opt.MCP.Server == nil {
		panic(errors.New("mcp server is nil"))
	}
	s := &MCPSrv{
		opt:    opt,
		server: faimcpserver.New(ctx, opt.MCP.Server, log.SysLog(), "Ut_Ai_Agent_MCP_Server", "1.0.0.1"),
		vp:     vp,
		vwr:    vwr,
		ext:    dal.Ext(),
		db:     dal.Db(),
	}
	k, err := s.ext.Knowledge()
	if err != nil {
		panic(err)
	}
	mcpTools, err := k.GetFontMCPTools(ctx, Knowledgem.GetFontMCPToolsReq{})
	if err != nil {
		return nil
	}
	if len(mcpTools.Data) > 0 {
		s.AddFontBusOperateTools(mcpTools.Data)
	} else {
		log.SysLog().Infof("font MCPTool is empty")
	}
	// 构建页面跳转工具
	s.AddPageRedirectionWithinTheSystemTool()
	// 构建联网搜索工具
	s.AddOpenWebSearchMcp(ctx)
	return s
}

type MCPSrv struct {
	opt    *config.Config
	server faimcpserver.If                                                    // MCP服务端
	vp     fmsg.PublishFunc                                                   // 消息发送器
	vwr    func(operateID string, waitTime time.Duration) (*fmsg.UMsg, error) // 消息等待器
	ext    ext.If                                                             // 数据访问层
	db     db.If
}

func (s *MCPSrv) Publish(topic string, msg *msgm.TAiAgentMessage) {
	if s.vp == nil {
		return
	}
	uMsg := fmsg.NewUMsg(topic, msg, "McpServer", []string{fmsg.ToMqtt}, nil)
	log.SysLog().Infof("server publish msg: %v", uMsg.String("推送到消息总线"))
	s.vp(uMsg)
}

func (s *MCPSrv) AddPageRedirectionWithinTheSystemTool() {
	tinfo := s.server.NewTool("page_redirection_within_the_system_tool",
		mcp.WithDescription("This is a page redirection tool within the system that requires the provision of session ID and route Addr, both of which exist within the relevant context.Note: Before calling, please call the query_route_tool tool to obtain the routing address. There is no need to concatenate routes based on the hierarchical relationship of routeName. Simply take the routeAdd value of the nearest routeName"),
		mcp.WithString("sessionId",
			mcp.Description("Chat provided session ID"),
			mcp.Required(),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "页面跳转",
			ReadOnlyHint:    Of(true),
			DestructiveHint: Of(false),
			IdempotentHint:  Of(false),
			OpenWorldHint:   Of(true),
		}),
		mcp.WithString("routeAddr",
			mcp.Description("route addr"),
			mcp.Required(),
		),
	)

	s.server.AddTool(tinfo, func(ctx context.Context, request mcp.CallToolRequest) (res *mcp.CallToolResult, err error) {
		defer urecover.HandlerRecover(fmt.Sprintf("MCP工具[%s]调用异常", tinfo.Name), &err)
		sessionId, ok := request.GetArguments()["sessionId"].(string)
		if !ok {
			log.SysLog().Errorf("页面跳转失败,缺少sessionId参数")
			res = mcp.NewToolResultText("页面跳转失败,缺少sessionId参数")
			AddMeta(res, "", 2, "页面跳转失败,缺少sessionId参数")
			return res, nil
		}
		route, ok := request.GetArguments()["routeAddr"].(string)
		if !ok {
			log.SysLog().Errorf("页面跳转失败")
			res = mcp.NewToolResultText("页面跳转失败,缺少routeAddr参数")
			AddMeta(res, sessionId, 2, "页面跳转失败,缺少routeAddr参数")
			return mcp.NewToolResultText("页面跳转失败,缺少routeAddr参数"), nil
		}

		operateID := uuid.New().String()
		var data struct {
			Object struct {
				Url string `json:"url"`
			} `json:"object"`
		}
		data.Object.Url = route
		msg := msgm.TAiAgentMessage{
			SessionId:   sessionId,
			MessageType: msgm.MessageType_Page_Redirection,
			Topic:       iconsts.Topic_Bus_Page_Redirection_Within_The_System_Tool_Ca,
			RespType:    msgm.RespType_ServerReqArk,
			OperateID:   operateID,
			Data:        data,
		}
		s.Publish(iconsts.Topic_Mqtt_To_Font_Ca+"/"+sessionId, &msg)
		uRMsg, err := s.vwr(operateID, time.Second*60)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				res = mcp.NewToolResultText("页面跳转失败,调用超时")
				AddMeta(res, sessionId, 2, "页面跳转失败,调用超时")
				return res, nil
			}
			log.SysLog().Errorf("页面跳转失败,调用异常,请联系管理员处理：%s", err.Error())
			res = mcp.NewToolResultText("页面跳转工具,调用异常,请联系管理员处理")
			AddMeta(res, sessionId, 2, "页面跳转失败,调用异常")
			return res, nil
		} else {
			data1, ok1 := uRMsg.Msg.([]byte)
			if !ok1 {
				log.SysLog().Errorf("页面跳转失败,应答数据异常")
				res = mcp.NewToolResultText("页面跳转失败,应答数据异常")
				AddMeta(res, sessionId, 2, "页面跳转失败,应答数据异常")
				return res, nil
			}
			// 转换成应答消息
			var result struct {
				Message string `json:"message"`
				Status  int    `json:"status"`
			}
			rmsg := msgm.TAiAgentMessage{}
			rmsg.Data = &result
			err = json.Unmarshal(data1, &rmsg)
			if err != nil {
				log.SysLog().Errorf("页面跳转失败,应答数据结构异常:%s", err.Error())
				res = mcp.NewToolResultText("页面跳转失败,应答数据结构异常")
				AddMeta(res, sessionId, 2, "页面跳转失败,应答数据结构异常")
				return mcp.NewToolResultText("页面跳转失败,应答数据结构异常"), nil
			}
			res = mcp.NewToolResultText(result.Message)
			matemsg := ""
			if result.Status == 1 {
				matemsg = "页面跳转成功"
			} else {
				matemsg = "页面跳转失败," + result.Message
			}
			AddMeta(res, sessionId, result.Status, matemsg)
			return res, nil
		}
	})
}
func AddMeta(result *mcp.CallToolResult, sessionId string, status int, message string) {
	if result.Result.Meta == nil {
		result.Result.Meta = make(map[string]any)
	}
	result.Meta["sessionId"] = sessionId // 固定要发
	result.Meta["status"] = status       // int ->1:成功;2:失败
	result.Meta["message"] = message     // 例如：查询交换机ID，成功的内容（查询交换机ID成功），或失败的原因（异常：未发现该交换机）
	return
}
func (s *MCPSrv) AddFontBusOperateTools(tools []Knowledgem.McpToolInfo) {
	for _, t := range tools {
		var isRequired func(schema map[string]any)
		properties := make(map[string]any, 0)
		if t.Object != nil {
			isRequired = func(schema map[string]any) {
				schema["required"] = t.Object.IsRequired
			}
			for _, p := range t.Object.Properties {
				if p.Type == "array" && p.ArrayItem != nil {
					items := map[string]any{
						"type":        p.ArrayItem.Type,
						"description": p.ArrayItem.Des,
					}
					properties[p.Name] = map[string]any{
						"type":        p.Type,
						"description": p.Des,
						"items":       items,
					}
				} else {
					properties[p.Name] = map[string]any{
						"type":        p.Type,
						"description": p.Des,
					}
				}
			}
		}
		var ti mcp.Tool
		if t.Object != nil {
			var prop func(schema map[string]any)
			prop = func(schema map[string]any) {
				if len(properties) > 0 {
					schema["properties"] = properties
				}
			}
			if t.Annotations.Title == "" {
				t.Annotations.Title = t.Name
			}
			ti = s.server.NewTool(
				t.Name,
				mcp.WithDescription(t.Des),
				mcp.WithString("sessionId",
					mcp.Description("Chat provided session ID"),
					mcp.Required(),
				),
				mcp.WithObject("object",
					isRequired,
					prop,
					mcp.Description(t.Object.Des),
				),
				mcp.WithToolAnnotation(t.Annotations),
			)
		} else {
			ti = s.server.NewTool(
				t.Name,
				mcp.WithDescription(t.Des),
				mcp.WithString("sessionId",
					mcp.Description("Chat provided session ID"),
					mcp.Required(),
				),
				mcp.WithToolAnnotation(t.Annotations),
			)
		}
		s.server.AddTool(ti, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			defer urecover.HandlerRecover(fmt.Sprintf("MCP工具[%s]调用异常", ti.Name), nil)
			sessionId, ok := request.GetArguments()["sessionId"].(string)
			if !ok {
				log.SysLog().Errorf("%s失败,入参缺少sessionId", t.Annotations.Title)
				res := mcp.NewToolResultText(fmt.Sprintf("%s失败,入参缺少sessionId", t.Annotations.Title))
				// 注入应答信息
				AddMeta(res, sessionId, 2, fmt.Sprintf("%s失败,入参缺少sessionId", t.Annotations.Title))
				return res, nil
			}
			delete(request.GetArguments(), "sessionId")
			operateID := utils.GetStringID()
			msg := msgm.TAiAgentMessage{
				SessionId:   sessionId,
				MessageType: msgm.MessageType_Bus_Operate,
				Topic:       ti.Name,
				RespType:    msgm.RespType_ServerReqArk,
				OperateID:   operateID,
				Data:        request.Params.Arguments,
			}
			s.Publish(iconsts.Topic_Mqtt_To_Font_Ca+"/"+sessionId, &msg)
			uRMsg, err := s.vwr(operateID, time.Second*60)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					log.SysLog().Errorf("%s失败,执行超时", t.Annotations.Title)
					res := mcp.NewToolResultText(fmt.Sprintf("%s失败,执行超时", t.Annotations.Title))
					// 注入应答信息
					AddMeta(res, sessionId, 2, fmt.Sprintf("%s失败,执行超时", t.Annotations.Title))
					return res, nil
				}
				log.SysLog().Errorf("%s失败,系统异常：%s", t.Annotations.Title, err.Error())
				res := mcp.NewToolResultText(fmt.Sprintf("%s失败,系统异常", t.Annotations.Title))
				// 注入应答信息
				AddMeta(res, sessionId, 2, fmt.Sprintf("%s失败,系统异常", t.Annotations.Title))
				return res, nil
			} else {
				bodyBuf, ok1 := uRMsg.Msg.([]byte)
				if !ok1 {
					log.SysLog().Errorf(fmt.Sprintf("%s失败,应答数据异常", t.Annotations.Title))
					res := mcp.NewToolResultText(fmt.Sprintf("%s失败,系统异常", t.Annotations.Title))
					// 注入应答信息
					AddMeta(res, sessionId, 2, fmt.Sprintf("%s失败,系统异常", t.Annotations.Title))
					return res, nil
				}
				// 转换成应答消息
				var result struct {
					Message string `json:"message"`
					Status  int    `json:"status"`
				}
				RecMsg := msgm.TAiAgentMessage{}
				RecMsg.Data = &result
				err = json.Unmarshal(bodyBuf, &RecMsg)
				if err != nil {
					log.SysLog().Errorf("%s失败,应答数据结构异常:%s", t.Annotations.Title, err.Error())
					res := mcp.NewToolResultText(fmt.Sprintf("%s失败,应答数据结构异常", t.Annotations.Title))
					// 注入应答信息
					AddMeta(res, sessionId, 2, fmt.Sprintf("%s失败,应答数据结构异常", t.Annotations.Title))
					return res, nil
				}
				res := mcp.NewToolResultText(result.Message)
				mateMsg := ""
				if result.Status == 1 {
					mateMsg = t.Annotations.Title + "成功"
				} else {
					mateMsg = t.Annotations.Title + "失败," + result.Message
				}
				// 注入应答信息
				AddMeta(res, sessionId, result.Status, mateMsg)
				return res, nil
			}
		})
	}
}
func (s *MCPSrv) AddOpenWebSearchMcp(ctx context.Context) {
	// 先暂时固化到开发服务器
	mcpClien, _ := faimcpclient.New(ctx, "openwebsearch", &faimcp.Option{
		Type:        "streamable",
		Command:     "",
		Env:         nil,
		Args:        nil,
		BaseURL:     "http://192.168.53.217:3000/mcp",
		Header:      nil,
		OAuthConfig: nil,
	}, log.SysLog())
	ti := s.server.NewTool(
		"general_tool_internet_search_tool",
		mcp.WithDescription("This is a general tool for Internet search"),
		mcp.WithString("sessionId",
			mcp.Description("Chat provided session ID"),
			mcp.Required(),
		),
		mcp.WithString("query",
			mcp.Description("Search query must not be empty"),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "联网搜索",
			ReadOnlyHint:    Of(false),
			DestructiveHint: Of(false),
			IdempotentHint:  Of(false),
			OpenWorldHint:   Of(true),
		}),
	)
	s.server.AddTool(
		ti,
		func(ctx context.Context, request mcp.CallToolRequest) (res *mcp.CallToolResult, err error) {
			defer urecover.HandlerRecover(fmt.Sprintf("MCP工具[%s]调用异常", ti.Name), &err)
			sessionId, ok := request.GetArguments()["sessionId"].(string)
			if !ok {
				log.SysLog().Errorf("联网搜索失败,入参缺少sessionId")
				res = mcp.NewToolResultText("联网搜索失败,入参缺少sessionId")
				// 注入应答信息
				AddMeta(res, sessionId, 2, "联网搜索失败,入参缺少sessionId")
				return res, nil
			}
			query, ok := request.GetArguments()["query"].(string)
			if !ok {
				log.SysLog().Errorf("联网搜索失败,入参缺少query")
				res = mcp.NewToolResultText("联网搜索失败,入参缺少query")
				// 注入应答信息
				AddMeta(res, sessionId, 2, "联网搜索失败,入参缺少query")
				return res, nil
			}
			request.Params.Name = "search"
			request.Params.Arguments = map[string]any{
				"query":   query,
				"engines": []string{"baidu"},
				"limit":   20,
			}
			callTool, err := mcpClien.CallTool(ctx, request)
			if err != nil || callTool.IsError {
				log.SysLog().Errorf("联网搜索失败,服务异常:%v", err)
				res = mcp.NewToolResultText(fmt.Sprintf("联网搜索失败,服务异常"))
				// 注入应答信息
				AddMeta(res, sessionId, 2, fmt.Sprintf("联网搜索失败,服务异常"))
				return res, nil
			}
			AddMeta(callTool, sessionId, 1, "联网搜索成功")
			return callTool, nil
		},
	)
}

var _ If = &MCPSrv{}

func Of[T any](v T) *T {
	return &v
}
