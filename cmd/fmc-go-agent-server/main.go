package main

import (
	"embed"
	"fmt"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/config"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/controller"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/router"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/log"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/middleware"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/urecover"
	"github.com/freedqo/fmc-go-agent/pkg/knife4jgo"
	"github.com/freedqo/fmc-go-agent/pkg/uconfvp"
	"github.com/freedqo/fmc-go-agent/pkg/webapp"
	"github.com/gin-gonic/gin"
	"os"
	"time"
)

// 需要嵌入的文件
//
//go:embed docs
var docsDir embed.FS

//	@title			fmc-go-agent服务Api文档
//	@version		1.0
//	@description	fmc-go-agent 服务Api文档
//	@contact.name	fmc-go-agent
//	@contact.url	https://github.com/freedqo/fmc-go-agent
//	@contact.email	2703155823@qq.cn
//	@host			http://localhost:7589
func main() {
	defer urecover.HandlerRecover("服务异常退出", nil)
	//解析命令行参数 -v
	FlagVersion()

	// 应用配置文件
	appCfg := config.NewDefault()
	confVp, err := uconfvp.New(appCfg, config.AppName)
	if err != nil {
		panic(err)
	}
	// 回写配置文件，更新退出时间
	defer func() {
		appCfg.Base.LastExitTime = time.Now().String()
		err = confVp.Write()
		if err != nil {
			log.SysLog().Errorf("配置文件写入异常：%s", err)
		}
	}()
	// 初始化日志
	log.NewLog(appCfg.Log)
	defer func() {
		err = log.Sync() //确保所有缓冲的日志都被写入
		if err != nil {
			os.WriteFile("./Panic.log", []byte(fmt.Sprintf("日志写入异常：%s", err)), 0644)
		}
	}()
	log.SysLog().Infof("程序启动,上次启动时间：%s", appCfg.Base.LastExitTime)
	// 构建应用
	app := webapp.New(config.AppName, log.SysLog())

	// 设置Gin为开发模式（默认）
	if Model != "release" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	// 实例Gin
	g := gin.New()
	g.MaxMultipartMemory = 8 << 20 // 8 MiB
	// 注册中间件
	//Recovery 中间件，用于恢复任何异常
	g.Use(middleware.GinRecoveryMiddleware())
	//Logger 中间件，用于记录请求日志
	g.Use(middleware.GinLoggerMiddleware())

	// 开发、调试模式,支持接口文档
	if Model != "release" {
		//注册接口文档 gin+swag+knife4jgo+knife4j(vue2)
		err := knife4jgo.K.SetServicesJsonFile(&docsDir)
		if err != nil {
			panic(err)
		}
		err = knife4jgo.K.SetSwaggerJsonFile(&docsDir)
		if err != nil {
			panic(err)
		}
		g.GET("/knife4jgo/*any", knife4jgo.K.GinKnife4jGoHandler)
		g.NoRoute(knife4jgo.K.GinKnife4jGoNoRouteHandler)
	}

	// 实例服务
	s := service.New(app.GetCtx(), appCfg)

	// 实例控制器
	c := controller.New(s)

	// 注册路由
	router.BindRoute(g, c)

	// 启动Http服务(Gin编码)
	app.ListenHttp(appCfg.Base.HttpPort, g)

	// 启动监控服务
	app.ListenMonitor(s.Start)

	// 进程监听
	app.Wait()

}
