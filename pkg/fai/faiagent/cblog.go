package faiagent

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

func newCbLog(log *zap.SugaredLogger) callbacks.Handler {
	// 获取底层 logger 以检查日志级别
	baseLogger := log.Desugar()
	levelEnabler := baseLogger.Core().Enabled

	builder := callbacks.NewHandlerBuilder()

	builder.OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
		// 记录开始信息（信息级别日志）
		log.Infof("[view]: start [%s:%s:%s]", info.Component, info.Type, info.Name)

		// 使用日志级别替代 config.Detail 判断
		if levelEnabler(zap.DebugLevel) {
			var b []byte
			var err error
			if levelEnabler(zap.DebugLevel) {
				b, err = json.MarshalIndent(input, "", "  ")
			} else {
				b, err = json.Marshal(input)
			}

			if err != nil {
				log.Error("Failed to marshal input", zap.Error(err))
				return ctx
			}
			log.Debugf("Callback input: %s", string(b))
		}
		return ctx
	})
	builder.OnStartWithStreamInputFn(func(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
		// 记录开始信息（信息级别日志）
		log.Debugf("[view]: StartWithStream Input [%s:%s:%s]", info.Component, info.Type, info.Name)

		return ctx
	})
	builder.OnEndWithStreamOutputFn(func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
		// 记录结束信息（信息级别日志）
		log.Debugf("[view]: EndWithStream input [%s:%s:%s]", info.Component, info.Type, info.Name)
		// 使用日志级别替代 config.Detail 判断
		if levelEnabler(zap.DebugLevel) {
			js, err := json.Marshal(output)
			if err == nil {
				log.Debugf("[view]: EndWithStream Output [%s]", js)
			}
		}
		return ctx
	})
	builder.OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
		// 记录错误信息（错误级别日志）
		log.Errorf("[view]: error [%s:%s:%s] %s", info.Component, info.Type, info.Name, err.Error())
		return ctx
	})

	builder.OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
		// 记录结束信息（信息级别日志）
		log.Infof("[view]: end [%s:%s:%s]", info.Component, info.Type, info.Name)

		// 仅在 Debug 级别启用时记录输出详情
		if levelEnabler(zap.DebugLevel) {
			b, err := json.Marshal(output)
			if err != nil {
				log.Error("Failed to marshal output", zap.Error(err))
				return ctx
			}
			log.Debugf("Callback output: %s", string(b))
		}
		return ctx
	})

	return builder.Build()
}
