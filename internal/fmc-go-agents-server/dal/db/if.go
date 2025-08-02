package db

import (
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/dal/db/dbif"
)

type If interface {
	Gdb() dbif.If
}
