package openaisrv

import "github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/service/openaisrv/v1srv"

type If interface {
	V1() v1srv.If
}
