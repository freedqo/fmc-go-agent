package utils

import (
	"os"
	"unicode"
)

func FirstToUpper(str string) string {
	if len(str) == 0 {
		return str
	}
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// FileExists 判断文件是否存在
// 返回：存在-true，不存在-false，错误（如权限问题）
func FileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil // 文件存在
	}
	if os.IsNotExist(err) {
		return false, nil // 文件不存在
	}
	return false, err // 其他错误（如权限不足）
}
func JoinXdDirPath(paths ...string) string {
	path := ""
	for _, v := range paths {
		if len(v) > 1 && v[len(v)-1] != '/' {
			path += v + "/"
		} else {
			path += v
		}
	}
	return path
}
func JoinXdFilePath(paths ...string) string {
	path := ""
	for _, v := range paths {
		if len(v) >= 1 && v[len(v)-1] != '/' {
			path += v + "/"
		} else {
			path += v
		}
	}
	path = path[:len(path)-1]
	return path
}

