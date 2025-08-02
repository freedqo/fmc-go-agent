package crontask

import (
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func NewCronLog(log *zap.SugaredLogger) cron.Logger {
	newLog := log.Desugar().WithOptions(zap.AddCallerSkip(1))
	return &CronLog{
		zapLog: newLog.Sugar(),
	}
}

type CronLog struct {
	zapLog *zap.SugaredLogger
}

func (c *CronLog) Info(msg string, keysAndValues ...interface{}) {
	c.zapLog.Infow(msg, keysAndValues...)
}

func (c *CronLog) Error(err error, msg string, keysAndValues ...interface{}) {
	c.zapLog.Errorw(msg, keysAndValues...)
}

var _ cron.Logger = &CronLog{}
