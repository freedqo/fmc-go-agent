package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	Model   = "debug"             //debug:开发、调试模式; release:生产模式(发版模式)
	Version = "开发测试版本（请勿部署至生产环境）" //版本号 V1.00.12.01
	MainGit = "主Git信息未知"          //主仓库的git信息
	DepGits = "依赖库Git信息未知"        //依赖库的git信息
)

func FlagVersion() {
	//解析命令行参数 -v
	var isShowVer = false
	flag.BoolVar(&isShowVer, "v", isShowVer, fmt.Sprintf("显示程序版本，用法举例：-v"))
	flag.Parse()
	if isShowVer {
		fmt.Println(GetVersion())
		os.Exit(0)
	}
}

func GetVersion() string {
	gitStr := strings.ReplaceAll(MainGit, "<br>", "\r\n")
	gitStr = strings.ReplaceAll(gitStr, "&nbsp", " ")
	depgitsStr := strings.ReplaceAll(DepGits, "<br>", "\r\n")
	depgitsStr = strings.ReplaceAll(depgitsStr, "&nbsp", " ")
	return fmt.Sprintf(
		"运行模式:%s"+
			"\r\n"+
			"版 本 号:%s"+
			"\r\n"+
			"%s"+
			"%s",
		Model,
		Version,
		gitStr,
		depgitsStr)
}
