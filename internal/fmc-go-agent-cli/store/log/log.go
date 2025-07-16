package log

import (
	"errors"
	"github.com/freedqo/fmc-go-agent/pkg/uzlog"
	kit_log "github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

var (
	logMaps = make(map[string]*uzlog.UzLog)
	sysLog  *uzlog.UzLog // 系统日志
	kitLog  kit_log.Logger
	once    sync.Once
)

// NewLog 实例化系统日志
// 入参: config *sysconfig.SysLocalConfigFile 系统配置文件
// 返回: 无
func NewLog(opt *uzlog.Option) {
	var c *uzlog.Option
	once.Do(func() {
		if opt == nil {
			c = uzlog.NewDefaultOption()
		} else {
			c = opt
		}
		err := c.Verify()
		if err != nil {
			panic(err)
		}
		c.Name = "cli"
		sysLog = uzlog.NewUzLog(c)

	})
	if sysLog == nil {
		panic("sysLog is nil")
	}

	logMaps[sysLog.Name()] = sysLog
	return
}

// Sync 同步日志
// 入参: 无
// 返回: error 错误信息
func Sync() error {
	err := sysLog.Sync()
	if err != nil {
		return err
	}

	return nil
}

// SysLog 获取系统日志实例
// 入参: 无
// 返回: *zap.SugaredLogger 日志实例
func SysLog() *zap.SugaredLogger {
	if sysLog == nil {
		NewLog(nil)
	}
	return sysLog.Sugar()
}

func GetLogLevel(name string) (string, zapcore.Level, error) {
	v, ok := logMaps[name]
	if !ok {
		return "", -1, errors.New("not found log name")
	}
	return v.Name(), v.GetLevel(), nil
}
func SetLogLevel(name string, level zapcore.Level) error {
	v, ok := logMaps[name]
	if !ok {
		return errors.New("not found log name")
	}
	v.SetLevel(level)
	return nil
}
func GetAllLogLevel() (map[string]zapcore.Level, error) {
	mp := make(map[string]zapcore.Level, 0)
	for k, v := range logMaps {
		mp[k] = v.GetLevel()
	}
	return mp, nil
}
func SetAllLogLevel(level zapcore.Level) error {
	for _, v := range logMaps {
		v.SetLevel(level)
	}
	return nil
}

func KitLog() kitlog.Logger {
	return sysLog
}
