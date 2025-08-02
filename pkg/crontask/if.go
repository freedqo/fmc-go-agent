package crontask

import (
	"github.com/freedqo/fmc-go-agents/pkg/commif"
)

type If interface {
	GetTask(name string) (*Task, bool)
	GetTaskList() []*Task
	Add(name, schedule string, job func()) error
	Delete(name string) error
	Update(name, newSchedule string) error
	StopTask(name string) error
	commif.MonitorIf
}
