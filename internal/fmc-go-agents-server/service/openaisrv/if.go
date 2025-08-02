package openaisrv

import "github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/service/openaisrv/v1srv"

type If interface {
	V1() v1srv.If
}
