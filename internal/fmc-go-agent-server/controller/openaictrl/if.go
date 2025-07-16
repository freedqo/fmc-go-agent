package openaictrl

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/controller/openaictrl/v1ctrl"
)

type If interface {
	V1() v1ctrl.If
}
