package crontask

import (
	"context"
	"fmt"
	"github.com/freedqo/fmc-go-agent/pkg/utils"
	"go.uber.org/zap"
	"sync"
)

// New 创建一个新的CronTask实例
// 入参： log *zap.SugaredLogger 日志实例
// 返回： If CronTask实例
func New(log *zap.SugaredLogger) If {
	clog := NewCronLog(log)
	ct := &CronTask{
		tasks: make(map[string]*Task),
		mu:    sync.Mutex{},
		log:   log,
	}
	ct.cron = cron.New(
		cron.WithSeconds(),
		cron.WithLogger(clog),
		cron.WithChain(ct.recover()),
		cron.WithChain(cron.SkipIfStillRunning(clog)),
	)
	return ct
}

type Task struct {
	Name        string       // 任务名称
	Schedule    string       // 定时任务表达式
	job         func()       // 任务
	scheduledID cron.EntryID // 记录任务ID
	reqRun      bool         // 需要启动
}

type CronTask struct {
	ctx    context.Context
	cancel context.CancelFunc
	cron   *cron.Cron
	tasks  map[string]*Task
	mu     sync.Mutex // 用于确保对任务的操作是线程安全的
	log    *zap.SugaredLogger
}

// GetTask 根据名称获取任务
// 入参： name string 任务名称
// 返回： *Task 任务实例
// 返回： bool 任务是否存在
func (tm *CronTask) GetTask(name string) (*Task, bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	task, exists := tm.tasks[name]
	return task, exists
}

// GetTaskList 获取任务列表
// 入参： 无
// 返回： []*Task 任务列表
func (tm *CronTask) GetTaskList() []*Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	list := make([]*Task, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		list = append(list, task)
	}
	return list
}

// Add 添加一个新的任务
// 入参： name string 任务名称
// 入参： schedule string 定时任务表达式
// 返回： error 错误信息
func (tm *CronTask) Add(name, schedule string, job func()) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tasks[name]; exists {
		return fmt.Errorf("task with Name %s already exists", name)
	}

	tm.tasks[name] = &Task{
		Name:     name,
		Schedule: schedule,
		job:      job,
		reqRun:   true,
	}
	return nil
}

// Delete 删除一个任务
// 入参： name string 任务名称
// 返回： error 错误信息
func (tm *CronTask) Delete(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task, exists := tm.tasks[name]
	if !exists {
		return fmt.Errorf("task with Name %s does not exist", name)
	}
	if task.reqRun {
		tm.cron.Remove(task.scheduledID)
	}
	delete(tm.tasks, name)

	return nil
}

// Update 修改任务的执行时间
// 入参： name string 任务名称
// 入参： newSchedule string 新的定时任务表达式
// 返回： error 错误信息
func (tm *CronTask) Update(name, newSchedule string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task, exists := tm.tasks[name]
	if !exists {
		return fmt.Errorf("task with Name %s does not exist", name)
	}
	task.Schedule = newSchedule

	if task.reqRun {
		// 移除旧的任务
		tm.cron.Remove(task.scheduledID)
		task.reqRun = false
		// 添加新的任务
		entryID, err := tm.cron.AddFunc(task.Schedule, task.job)
		if err != nil {
			return err
		}
		task.scheduledID = entryID
	}
	return nil
}

// StarTask 停止任务管理器
// 入参： name string 任务名称
// 返回： error 错误信息
func (tm *CronTask) StarTask(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	task, exists := tm.tasks[name]
	if !exists {
		return fmt.Errorf("task with Name %s does not exist", name)
	}
	if task.reqRun {
		return nil
	}
	entryID, err := tm.cron.AddFunc(task.Schedule, task.job)
	if err != nil {
		return err
	}
	task.reqRun = true
	task.scheduledID = entryID
	return nil
}

// StopTask 停止任务管理器
// 入参： name string 任务名称
// 返回： error 错误信息
func (tm *CronTask) StopTask(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	task, exists := tm.tasks[name]
	if !exists {
		return fmt.Errorf("task with Name %s does not exist", name)
	}
	if !task.reqRun {
		return nil
	}
	tm.cron.Remove(task.scheduledID)
	task.reqRun = false
	task.scheduledID = cron.EntryID(-1)
	return nil
}

// Start 启动任务管理器
// 入参： ctx context.Context 上下文
// 返回： done <-chan struct{} 任务完成信号
// 返回： error 错误信息
func (tm *CronTask) Start(ctx context.Context) (done <-chan struct{}, err error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	d := make(chan struct{})
	lctx, cancel := context.WithCancel(ctx)
	tm.ctx = ctx
	tm.cancel = cancel
	for _, v := range tm.tasks {
		if !v.reqRun {
			continue
		}
		entryID, err := tm.cron.AddFunc(v.Schedule, v.job)
		if err != nil {
			return nil, err
		}
		v.scheduledID = entryID
	}
	tm.cron.Start()
	go func() {
		defer func() {
			tm.cron.Stop()
			for _, v := range tm.tasks {
				if v.reqRun {
					tm.cron.Remove(v.scheduledID)
				}
			}
			close(d)
		}()

		<-lctx.Done()
	}()
	return d, nil
}

// Stop 停止任务管理器
// 入参： 无
// 返回： error 错误信息
func (tm *CronTask) Stop() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if tm.cancel == nil {
		return nil
	}
	tm.cancel()
	return nil
}

// RestStart 重启任务管理器
// 入参： 无
// 返回： done <-chan struct{} 任务完成信号
// 返回： error 错误信息
func (tm *CronTask) RestStart() (done <-chan struct{}, err error) {
	err = tm.Stop()
	if err != nil {
		return nil, err
	}
	return tm.Start(tm.ctx)
}

var _ If = &CronTask{}

// recover 恢复
// 入参： 无
// 返回： cron.JobWrapper 恢复函数
func (tm *CronTask) recover() cron.JobWrapper {
	return func(j cron.Job) cron.Job {
		return cron.FuncJob(func() {
			defer func() {
				if r := recover(); r != nil {
					tm.log.Errorf("Cron定时任务调度器发生异常,Panic:%v,Stack:%s", r, utils.StackSkip(1, -1))
				}
			}()
			j.Run()
		})
	}
}
