package utils

import (
	"runtime"
	"strings"
)

// Stack returns the stack trace as a string.
func Stack() string {
	var buf [4096 * 2]byte
	n := runtime.Stack(buf[:], false)
	return string(buf[:n])
}

// StackSkip returns the stack trace as a string with configurable depth.
// skip: 跳过的栈帧数 (0 表示当前函数)
// maxFrames: 最大捕获的帧数 (-1 表示不限制)
func StackSkip(skip int, maxFrames int) string {
	// 跳过 Stack 函数自身和调用 Stack 的函数
	skip += 2

	var (
		buf       []byte
		stackSize int
	)

	// 动态调整缓冲区大小，直到足够大
	for size := 4096; ; size *= 2 {
		buf = make([]byte, size)
		stackSize = runtime.Stack(buf, false)
		if stackSize < size {
			break
		}
	}

	// 解析堆栈跟踪
	frames := strings.Split(string(buf[:stackSize]), "\n")
	if len(frames) == 0 {
		return ""
	}

	// 应用跳过和最大帧数限制
	var result []string
	frameCount := 0

	for i := 0; i < len(frames); i += 2 {
		if i/2 < skip {
			continue // 跳过指定的帧数
		}

		if maxFrames >= 0 && frameCount >= maxFrames {
			break // 达到最大帧数限制
		}

		if i+1 < len(frames) {
			result = append(result, frames[i], frames[i+1])
			frameCount++
		}
	}

	return strings.Join(result, "\n")
}

// StackWithLevel 返回指定层级深度的堆栈跟踪
// level: 堆栈层级深度 (1 表示调用者，2 表示调用者的调用者，依此类推)
func StackWithLevel(level int) string {
	return StackSkip(level, 10) // 默认捕获最多10帧
}

// StackAll 返回完整的堆栈跟踪
func StackAll() string {
	return StackSkip(0, -1) // 不跳过任何帧，捕获全部
}
