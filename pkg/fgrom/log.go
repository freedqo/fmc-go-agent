package fgrom

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
	"time"
)

func NewGormLog(log *zap.SugaredLogger) logger.Interface {
	newLog := log.Desugar().WithOptions(zap.AddCallerSkip(1), zap.IncreaseLevel(zap.WarnLevel))
	return &GormLogger{
		log: newLog.Sugar(),
	}
}

// 慢 SQL 阈值，单位为毫秒
const slowSQLThreshold = 1 * 1000

// GormLogger 自定义 gorm 日志器
type GormLogger struct {
	log *zap.SugaredLogger
}

func (g GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	var l zapcore.Level
	if level == logger.Info {
		l = zap.InfoLevel
	} else if level == logger.Warn {
		l = zap.WarnLevel
	} else if level == logger.Error {
		l = zap.ErrorLevel
	} else if level == logger.Silent { // 日志级别为 Silent 时，不记录日志
		l = 100
	}
	g.log = g.log.Desugar().WithOptions(zap.IncreaseLevel(l)).Sugar()
	return g
}

func (g GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	g.log.Infof(s, i...)
}

func (g GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	g.log.Warnf(s, i...)
}

func (g GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	g.log.Errorf(s, i...)
}

func (g GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		g.log.Errorf("SQL error: %v, SQL: %s, Rows: %d, Elapsed: %v", err, sql, rows, elapsed.String())
	} else if elapsed > time.Duration(slowSQLThreshold)*time.Millisecond {
		g.log.Warnf("Slow SQL executed: %s, Rows: %d, Elapsed: %v", sql, rows, elapsed.String())
	}
}

var _ logger.Interface = &GormLogger{}
