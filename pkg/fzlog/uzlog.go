package fzlog

import (
	"errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
)

// NewUzLog 创建一个zap日志
// 入参: cfg *LogConfig 日志配置
// 返回: *UzLog 日志对象
func NewUzLog(cfg *Option) *UzLog {
	//日志编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "Ts",
		LevelKey:       "Level",
		NameKey:        "Logger",
		CallerKey:      "Caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "Msg",
		StacktraceKey:  "Stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	// 处理空路径
	if cfg.Path == "" {
		root, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		cfg.Path = path.Join(root, "logs")
	}

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	//配置日志文件输出
	lumberjackLogger := &lumberjack.Logger{
		Filename:   path.Join(cfg.Path, cfg.Name+".log"), //日志名称
		MaxSize:    cfg.MaxSize,                          // 文件最大Size 单位M
		MaxBackups: cfg.MaxBackups,                       // 最大备份数量
		MaxAge:     cfg.MaxAge,                           // 最大备份天数
		Compress:   cfg.Compress,                         // 是否压缩
		LocalTime:  cfg.LocalTime,                        // 是否使用本地时间
	}
	//同步输出控制台和日志文件
	wr := io.MultiWriter(lumberjackLogger, os.Stdout)
	writeSyncer := zapcore.AddSync(wr)
	//配置日志运行级别
	//new日志的core
	// 创建可动态调整级别的配置
	level := zap.NewAtomicLevel()
	level.SetLevel(zapcore.Level(cfg.Level)) // 设置初始级别
	core := zapcore.NewCore(encoder, writeSyncer, level)
	log := zap.New(core, zap.AddCaller())
	zapLog := &UzLog{
		atomicLevel: level,
		Logger:      log,
		cfg:         cfg,
		name:        cfg.Name,
	}
	return zapLog
}

type UzLog struct {
	*zap.Logger
	atomicLevel zap.AtomicLevel
	name        string
	cfg         *Option
}

func (l *UzLog) SugaredLogger() *zap.SugaredLogger {
	return l.Logger.Sugar()
}

func (l *UzLog) Name() string {
	return l.name
}

func (l *UzLog) GetLevel() zapcore.Level {
	return l.atomicLevel.Level()
}

func (l *UzLog) SetLevel(level zapcore.Level) {
	l.atomicLevel.SetLevel(level)
}

func (l *UzLog) Log(keyvals ...interface{}) error {
	if len(keyvals)%2 != 0 {
		return errors.New("key-value参数必须成对出现")
	}
	// 将interface{}参数转换为zap字段
	fields := make([]zap.Field, 0, len(keyvals)/2)
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 >= len(keyvals) {
			return errors.New("key-value参数不完整")
		}
		key, ok := keyvals[i].(string)
		if !ok {
			return errors.New("键必须是字符串类型")
		}
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}
	// 记录日志
	l.Debug("KitLog记录", fields...)
	return nil
}
