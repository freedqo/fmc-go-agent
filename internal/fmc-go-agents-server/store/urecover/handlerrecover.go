package urecover

import (
	"errors"
	"fmt"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/store/log"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"github.com/google/uuid"
)

// HandlerRecover 捕获panic，并记录错误日志,必须在defer中调用，不能在defer func(){}中调用
func HandlerRecover(title string, inErr *error) {
	if err := recover(); err != nil {
		// 生成一个唯一的错误ID，用于后续的错误跟踪
		PanicId := uuid.New().ID()
		if inErr != nil {
			*inErr = errors.New(fmt.Sprintf("%s,遇到异常,PanicId: %d, Panic: %v", title, PanicId, err))
		}
		// 记录 panic 信息
		log.SysLog().Errorf("%s,遇到异常,PanicId: %d, Panic: %v, 堆栈: %v ,请联系管理员处理", title, PanicId, err, utils.StackSkip(1, -1))
	}
}
