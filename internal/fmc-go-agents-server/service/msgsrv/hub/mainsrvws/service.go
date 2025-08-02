package mainsrvws

import (
	"context"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/store/log"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg"
	"github.com/freedqo/fmc-go-agents/pkg/fmsg/uwss"
)

func New(cfg *uwss.Option) fmsg.MessageHubIf {
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
func (s *Service) GetConnectState() fmsg.ClientConnectState {
	return s.utWsServer.GetConnectState()
}

func (s *Service) FrontHandleMessage(msg *fmsg.UMsg) {
	//暂无需求，不处理

}

func (s *Service) Publish(msg *fmsg.UMsg) {
	s.utWsServer.Publish(msg)
}

func (s *Service) SubscribeRecEvent(fun *fmsg.RecEventFunc) {
	s.utWsServer.SubscribeRecEvent(fun)
}

func (s *Service) UnSubscribeRecEvent(fun *fmsg.RecEventFunc) {
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
