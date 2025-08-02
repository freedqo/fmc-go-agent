package openaictrl

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/controller/openaictrl/v1ctrl"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service"
)

func New(service service.If) If {
	return &Controller{
		service: service,
		v1:      v1ctrl.New(service),
	}
}

type Controller struct {
	service service.If
	v1      v1ctrl.If
}

func (c *Controller) V1() v1ctrl.If {
	return c.v1
}

var _ If = &Controller{}
