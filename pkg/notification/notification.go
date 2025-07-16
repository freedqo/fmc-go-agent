package notification

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Notification 表示一个通知
type Notification struct {
	ID      string      // 通知唯一标识
	Content interface{} // 通知内容
}

// Notifier 通知处理器，支持防抖功能
type Notifier struct {
	notificationChan chan Notification    // 通知队列
	stopChan         chan struct{}        // 停止信号channel
	timer            *time.Timer          // 定时器
	mutex            sync.Mutex           // 互斥锁
	debounceTime     time.Duration        // 防抖超时时间
	handler          func(Notification)   // 通知处理函数
	summaryHandler   func([]Notification) // 汇总通知处理函数
}

// NewNotifier 创建一个新的通知处理器
func NewNotifier(debounceTime time.Duration) *Notifier {
	return &Notifier{
		notificationChan: make(chan Notification, 1*1024),
		stopChan:         make(chan struct{}),
		debounceTime:     debounceTime,
	}
}

// SetHandler 设置通知处理函数
func (n *Notifier) SetHandler(handler func(Notification)) {
	n.handler = handler
}

// SetSummaryHandler 设置汇总通知处理函数
func (n *Notifier) SetSummaryHandler(handler func([]Notification)) {
	n.summaryHandler = handler
}

// Start 启动通知处理器
func (n *Notifier) Start(ctx context.Context) error {
	if n.handler == nil {
		return errors.New("通知处理函数未设置")
	}

	if n.summaryHandler == nil {
		return errors.New("汇总通知处理函数未设置")
	}

	go n.processNotifications(ctx)
	return nil
}

// Stop 停止通知处理器
func (n *Notifier) Stop() {
	close(n.stopChan)
	// 确保定时器停止
	n.mutex.Lock()
	if n.timer != nil {
		n.timer.Stop()
		n.timer = nil
	}
	n.mutex.Unlock()
}

// Push 向队列中添加通知
func (n *Notifier) Push(notification Notification) {
	select {
	case n.notificationChan <- notification:
		// 通知成功入队
	default:
		// 队列已满，这里选择丢弃，可根据需求修改为阻塞
	}
}

// processNotifications 处理通知的后台goroutine
func (n *Notifier) processNotifications(ctx context.Context) {
	var notifications []Notification
	var timerC <-chan time.Time

	for {
		select {
		case <-ctx.Done():
			// 上下文取消，退出
			return
		case <-n.stopChan:
			// 收到停止信号，处理剩余通知并退出
			n.mutex.Lock()
			if len(notifications) > 0 {
				n.summaryHandler(notifications)
				notifications = []Notification{}
			}
			n.mutex.Unlock()
			return
		case notification, ok := <-n.notificationChan:
			if !ok {
				// channel已关闭
				return
			}

			// 处理新通知
			n.mutex.Lock()
			notifications = append(notifications, notification)

			// 重置定时器
			if n.timer != nil {
				n.timer.Stop()
			}

			n.timer = time.AfterFunc(n.debounceTime, func() {
				n.mutex.Lock()
				defer n.mutex.Unlock()

				if len(notifications) > 0 {
					n.summaryHandler(notifications)
					notifications = []Notification{}
				}
			})

			timerC = n.timer.C
			n.mutex.Unlock()

			// 处理单个通知
			go n.handler(notification)

		case <-timerC:
			// 定时器触发，处理汇总通知
			n.mutex.Lock()
			if len(notifications) > 0 {
				n.summaryHandler(notifications)
				notifications = []Notification{}
			}
			n.mutex.Unlock()
		}
	}
}
