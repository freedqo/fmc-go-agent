package commif

import "context"

// MonitorIf 监控服务启动、停止、重启的接口
type MonitorIf interface {
	// Start 启动监控
	Start(ctx context.Context) (done <-chan struct{}, err error)
	// Stop 停止监控
	Stop() error
	// RestStart 重启监控
	RestStart() (done <-chan struct{}, err error)
}

