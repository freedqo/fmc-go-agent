package webapp

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

func New(name string, log *zap.SugaredLogger) *WebApp {
	ctx, cancel := context.WithCancel(context.Background())
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	return &WebApp{
		name:      name,
		wg:        sync.WaitGroup{},
		interrupt: interrupt,
		ctx:       ctx,
		cancel:    cancel,
		log:       log,
	}
}

type WebApp struct {
	name      string
	wg        sync.WaitGroup
	interrupt chan os.Signal
	ctx       context.Context
	cancel    context.CancelFunc
	log       *zap.SugaredLogger
	cfg       interface{}
}

func (a *WebApp) ListenHttp(port int, header *gin.Engine) {
	// 启动Http服务(Gin+Rpc)
	a.wg.Add(1)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: header,
	}
	go func() {
		defer a.wg.Done()
		defer a.log.Infof("关闭Http服务成功")
		go func() {
			// 启动Gin+Rpc服务
			a.log.Infof("Listening and serving HTTP on %s", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				a.log.Errorf("启动HttpServer监听服务异常: %s", err)
				return
			}
		}()
		<-a.ctx.Done()
		a.log.Infof("请求关闭Http服务")
		if err := srv.Shutdown(a.ctx); err != nil {
			a.log.Errorf("关闭Http服务异常:%s", err)
		}
	}()
}

func (a *WebApp) ListenTcpGobRpc(port int, header func(ctx context.Context, wg *sync.WaitGroup, conn net.Conn)) {
	a.wg.Add(1)
	tcp, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	go func() {
		defer a.wg.Done()
		defer a.log.Infof("关闭Tcp服务成功")
		tcpCWg := sync.WaitGroup{}
		go func() {
			a.log.Infof("Listening and serving TCP on %s", tcp.Addr())
			for {
				conn, err := tcp.Accept()
				if err != nil && !errors.Is(err, net.ErrClosed) {
					a.log.Errorf("基于Tcp Server的Rpc服务监听异常:%s", err)
					return
				}
				if conn != nil {
					tcpCWg.Add(1)
					go header(a.ctx, &tcpCWg, conn)
				}
			}
		}()
		<-a.ctx.Done()
		a.log.Infof("请求关闭Tcp服务")
		tcp.Close()
		tcpCWg.Wait()
	}()
}

func (a *WebApp) ListenMonitor(start func(ctx context.Context) (done <-chan struct{}, err error)) {
	if start == nil {
		return
	}
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		defer a.log.Infof("关闭后台通用监控服务成功")
		// 启动服务监听
		done, err := start(a.ctx)
		if err != nil {
			panic(err)
		}
		a.log.Infof("后台通用监控服务已启动")
		<-done
		a.interrupt <- os.Interrupt
	}()
}

func (a *WebApp) Wait() {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		<-a.interrupt
		a.log.Infof("请求停止服务")
		a.cancel()
	}()
	a.wg.Wait()
	a.log.Infof("服务已停止")
}

func (a *WebApp) GetCtx() context.Context {
	return a.ctx
}
