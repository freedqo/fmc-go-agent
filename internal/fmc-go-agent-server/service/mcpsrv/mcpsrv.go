package mcpsrv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antlabs/strsim"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/config"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal/ext"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/iconsts"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/dalm/extm/usercenterm"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/log"
	"github.com/freedqo/fmc-go-agent/pkg/uai/uaimcp/uaimcpserver"
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"time"
)

func New(ctx context.Context, opt *config.Config, ext ext.If, vp umsg.PublishFunc, vwr func(operateID string, waitTime time.Duration) (*umsg.UMsg, error)) If {
	s := &MCPSrv{
		mcp:        uaimcpserver.New(ctx, opt.McpServer, log.SysLog(), "Ut_Ai_Agent_MCP_Server", "1.0.0.1"),
		vp:         vp,
		vwr:        vwr,
		ext:        ext,
		routes:     make(map[string]*usercenterm.Route, 0),
		routeNames: make([]string, 0),
	}
	// 拉一下用户中心的缓存
	user, err := s.ext.UserCenter()
	if err != nil {
		panic(err)
	}
	// 登录用户中心
	login, err := user.Sm2Login(ctx, usercenterm.Sm2LoginReq{
		Account:  opt.Ext.UserCenterRoot.Username,
		Password: opt.Ext.UserCenterRoot.Password,
	})
	if err != nil {
		return nil
	}
	menuList, err := user.GetMenuList(ctx, usercenterm.GetMenuListReq{TokenId: login.Data.TokenID})
	if err != nil {
		return nil
	}
	routers := menuList.ToRoutes1()
	routers = append(routers, usercenterm.Route{
		RouteName: "页面跳转测试",
		RouteAddr: "AI/#/TestAiOpenPage",
	})
	for _, route := range routers {
		s.routes[route.RouteName] = &route
		s.routeNames = append(s.routeNames, route.RouteName)
	}
	//js, _ := json.Marshal(routers)
	//err = os.WriteFile("test.json", js, 0644)
	//if err != nil {
	//	panic(err)
	//}
	// 构建查询工具
	s.AddQueryRouteTool()
	// 构建页面跳转工具
	s.AddPageRedirectionWithinTheSystemTool()
	// 启动mcp服务
	_, err = s.mcp.Start(ctx)
	if err != nil {
		return nil
	}
	return s
}

type MCPSrv struct {
	mcp        uaimcpserver.If
	vp         umsg.PublishFunc
	vwr        func(operateID string, waitTime time.Duration) (*umsg.UMsg, error)
	ext        ext.If
	routes     map[string]*usercenterm.Route
	routeNames []string
}

func (s *MCPSrv) Publish(msg *umsg.Message) {
	if s.vp == nil {
		return
	}
	uMsg := umsg.NewUMsg(msg, "McpServer", []string{umsg.ToWsServer, umsg.ToMqtt})
	log.SysLog().Infof("mcp publish msg: %v", uMsg.String("推送到消息总线"))
	s.vp(uMsg)
}

func (s *MCPSrv) AddPageRedirectionWithinTheSystemTool() {
	tinfo := s.mcp.NewTool("page_redirection_within_the_system_tool",
		mcp.WithDescription("This is a page redirection tool within the system that requires the provision of session ID and route Addr, both of which exist within the relevant context.Note: Before calling, please call the query_route_tool tool to obtain the routing address. There is no need to concatenate routes based on the hierarchical relationship of routeName. Simply take the routeAdd value of the nearest routeName"),
		mcp.WithString("sessionId",
			mcp.Description("Chat provided session ID"),
			mcp.Required(),
		),
		mcp.WithString("routeAddr",
			mcp.Description("route addr"),
			mcp.Required(),
		),
	)
	s.mcp.AddTool(tinfo, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionId, ok := request.GetArguments()["sessionId"].(string)
		if !ok {
			log.SysLog().Errorf("missing sessionId")
			return nil, fmt.Errorf("missing sessionId")
		}
		route, ok := request.GetArguments()["routeAddr"].(string)
		if !ok {
			log.SysLog().Errorf("missing routeAddr")
			return nil, fmt.Errorf("missing routeAddr")
		}
		operateID := uuid.New().String()
		msg := umsg.Message{
			MessageBase: umsg.MessageBase{
				ClientID:        sessionId,
				Operate:         iconsts.Topic_Mqtt_Page_Redirection_Within_The_System_Tool_Ca,
				IsReplyOperate:  false,
				OperateID:       operateID,
				OperateDataType: "string",
			},
			OperateData: route,
		}
		s.Publish(&msg)
		uRMsg, err := s.vwr(operateID, time.Second*60)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return mcp.NewToolResultText("工具执行超时"), nil
			}
			return nil, err
		} else {
			data, ok1 := uRMsg.Msg.OperateData.(string)
			if ok1 {
				return mcp.NewToolResultText(data), nil
			} else {
				return nil, fmt.Errorf("工具执行失败,应答异常")
			}
		}
	})
}

func (s *MCPSrv) AddQueryRouteTool() {
	tinfo := s.mcp.NewTool("query_route_tool",
		mcp.WithDescription("This is a tool for querying routing information within the system"),
		mcp.WithString("routeName",
			mcp.Description("routeName,if routeName is an empty string, it can retrieve all menus (routes)"),
			mcp.Required(),
		),
	)
	s.mcp.AddTool(tinfo, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		routeName, ok := request.GetArguments()["routeName"].(string)
		if !ok {
			return nil, fmt.Errorf("missing sessionId")
		}
		if routeName == "" {
			data := make([]*usercenterm.Route, 0)
			for _, v := range s.routes {
				v1 := *v
				data = append(data, &v1)
			}
			if len(data) <= 0 {
				return mcp.NewToolResultText("系统未配置菜单参数"), nil
			}
			str, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			return mcp.NewToolResultText(string(str)), nil
		} else {
			log.SysLog().Infof("query routeName: %s", routeName)
			qRes := strsim.FindBestMatch(routeName, s.routeNames)
			data := make([]*usercenterm.Route, 0)
			for _, v := range qRes.AllResult {
				if v.Score >= 0.1 {
					data = append(data, s.routes[v.S])
				}
			}
			if len(data) <= 0 {
				for _, v := range s.routes {
					v1 := *v
					data = append(data, &v1)
				}
				str, err := json.Marshal(data)
				if err != nil {
					return nil, err
				}
				return mcp.NewToolResultText(string(str)), nil
			}
			str, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			return mcp.NewToolResultText(string(str)), nil
		}
	})
}

var _ If = &MCPSrv{}
