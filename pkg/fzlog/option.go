package fzlog

import (
	"fmt"
	"os"
	"path"
)

type Option struct {
	Name       string `comment:"日志器名称,系统覆盖,修改无效"`
	Level      int    `comment:"日志记录级别：-1->Debug及以上;0->Info及以上;1->Warn及以上;2->Error及以上;3->DPanic及以上;4->Panic及以上;5->Fatal及以上;"`
	MaxSize    int    `comment:"单个日志文件最大容量，单位:M,最小值1,最大值20"`
	MaxAge     int    `comment:"最大备份天数,最小值1,最大值1000"`
	MaxBackups int    `comment:"日志备份最大数量,最小值1,最大值1024"`
	LocalTime  bool   `comment:"是否使用本地时间"`
	Compress   bool   `comment:"是否压缩"`
	Path       string `comment:"日志文件存储路径(绝对路径)"`
}

func (c Option) Verify() error {
	if c.Level < -1 || c.Level > 5 {
		return fmt.Errorf("level must be between -1 and 5")
	}
	if c.MaxSize < 1 || c.MaxSize > 20 {
		return fmt.Errorf("MaxSize must be between 1 and 20")
	}
	if c.MaxAge < 1 || c.MaxAge > 1000 {
		return fmt.Errorf("MaxAge must be between 1 and 1000")
	}
	if c.MaxBackups < 1 || c.MaxBackups > 1024 {
		return fmt.Errorf("MaxBackups must be between 1 and 1024")
	}
	return nil
}

// NewDefaultOption 创建一个默认的日志配置
// 返回: *LogConfig 日志配置
func NewDefaultOption() *Option {
	root, err := os.Getwd()
	if err != nil {
		root = "./"
	}
	cfg := Option{
		Name:       "sysLog",
		Level:      0,
		MaxSize:    5,
		MaxAge:     30,
		MaxBackups: 1000,
		LocalTime:  true,
		Compress:   true,
	}
	if err != nil {
		cfg.Path = "./log"
	} else {
		cfg.Path = path.Join(root, "log")
	}
	return &cfg
}
