package rocketmqx

import (
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (r *RocketMQ) NewLog(name string, log *zap.SugaredLogger, level zapcore.LevelEnabler) rlog.Logger {
	newLog := log.Desugar().WithOptions(zap.AddCallerSkip(2), zap.IncreaseLevel(level))
	return &rocketMqLog{
		name: name,
		log:  newLog.Sugar(),
	}
}

type rocketMqLog struct {
	name string
	log  *zap.SugaredLogger
}

// Debug 实现 Debug 方法
func (r rocketMqLog) Debug(msg string, fields map[string]interface{}) {
	r.log.Debugw(fmt.Sprintf("RocketMq[%s] :%s ", r.name, msg), convertFields(fields)...)
}

// Info 实现 Info 方法
func (r rocketMqLog) Info(msg string, fields map[string]interface{}) {
	r.log.Infow(fmt.Sprintf("RocketMq[%s] :%s ", r.name, msg), convertFields(fields)...)
}

// Warning 实现 Warning 方法
func (r rocketMqLog) Warning(msg string, fields map[string]interface{}) {
	r.log.Warnw(fmt.Sprintf("RocketMq[%s] :%s ", r.name, msg), convertFields(fields)...)
}

// Error 实现 Error 方法
func (r rocketMqLog) Error(msg string, fields map[string]interface{}) {
	r.log.Errorw(fmt.Sprintf("RocketMq[%s] :%s ", r.name, msg), convertFields(fields)...)
}

// Fatal 实现 Fatal 方法
func (r rocketMqLog) Fatal(msg string, fields map[string]interface{}) {
	r.log.Fatalw(fmt.Sprintf("RocketMq[%s] :%s ", r.name, msg), convertFields(fields)...)
}

// Level 实现 Level 方法
func (r rocketMqLog) Level(level string) {
	// 不实现
}

// OutputPath 实现 OutputPath 方法
func (r rocketMqLog) OutputPath(path string) (err error) {
	//不实现了
	return nil
}

// convertFields 将 map[string]interface{} 转换为 zap 所需的 []interface{}
func convertFields(fields map[string]interface{}) []interface{} {
	var result []interface{}
	for key, value := range fields {
		result = append(result, key, value)
	}
	return result
}

var _ rlog.Logger = &rocketMqLog{}

