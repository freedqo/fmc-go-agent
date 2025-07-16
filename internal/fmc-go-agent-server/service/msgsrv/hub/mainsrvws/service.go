package mainsrvws

import (
	"context"
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/store/log"
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
	"github.com/freedqo/fmc-go-agent/pkg/umsg/uwss"
)

func New(cfg *uwss.Option) umsg.MessageHubIf {
	return &Service{
		Name:       "MainSrvWs",
		utWsServer: uwss.New("MainSrvWs", cfg, log.SysLog()),
	}
}

type Service struct {
	Name       string
	utWsServer uwss.If //utwebsocket服务端
}

func (s *Service) GetName() string {
	return s.Name
}
func (s *Service) GetConnectState() umsg.ClientConnectState {
	return s.utWsServer.GetConnectState()
}

func (s *Service) FrontHandleMessage(msg *umsg.UMsg) {
	//暂无需求，不处理

}

func (s *Service) Publish(msg *umsg.UMsg) {
	s.utWsServer.Publish(msg)
}

func (s *Service) SubscribeRecEvent(fun *umsg.RecEventFunc) {
	s.utWsServer.SubscribeRecEvent(fun)
}

func (s *Service) UnSubscribeRecEvent(fun *umsg.RecEventFunc) {
	s.utWsServer.UnSubscribeRecEvent(fun)
}

func (s *Service) Start(ctx context.Context) (done <-chan struct{}, err error) {
	return s.utWsServer.Start(ctx)
}

func (s *Service) Stop() error {
	return s.utWsServer.Stop()
}

func (s *Service) RestStart() (done <-chan struct{}, err error) {
	return s.utWsServer.RestStart()
}

var _ If = &Service{}
