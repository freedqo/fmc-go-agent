package db

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/dal/db/dbif"
)

type If interface {
	Gdb() dbif.If
}
