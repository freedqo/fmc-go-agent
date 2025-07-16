package build

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var Handler = &cli.Command{
	Name:  "build",
	Usage: "构建生产环境程序,输出win、linux的可执行文件",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "version",
			Aliases:  []string{"v"},
			Usage:    "指定程序版本号",
			Value:    "1.0.0",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "cmd",
			Aliases:  []string{"c"},
			Usage:    "指定cmd目录下的编译目录",
			Value:    "display-server-app",
			Required: true,
		},
		// 新增: 强制编译选项
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Usage:   "强制编译,即使存在未提交变更",
			Value:   false,
		},
	},
	Action: buildHandler,
}

func buildHandler(c *cli.Context) error {
	fmt.Println("[信息] 开始执行构建流程")
	startTime := time.Now()

	rootDir, err := findGoProjectRootDir()
	if err != nil {
		return fmt.Errorf("[错误] 查找项目根目录失败: %w", err)
	}
	fmt.Printf("[信息] 项目根目录已确定: %s\r\n", rootDir)

	// 获取命令行参数
	version := c.String("version")
	cmdDir := c.String("cmd")
	forceBuild := c.Bool("force") // 新增: 获取强制编译标志

	if strings.ReplaceAll(version, " ", "") == "" {
		return fmt.Errorf("[错误] 版本号不能为空")
	}
	fmt.Printf("[信息] 版本号已获取: %s\r\n", version)

	if strings.ReplaceAll(cmdDir, " ", "") == "" {
		return fmt.Errorf("[错误] cmd目录不能为空")
	}
	fmt.Printf("[信息] 编译目录已指定: %s\r\n", cmdDir)

	// 验证编译目录是否存在
	srcDir := filepath.Join(rootDir, "cmd", cmdDir)
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("[错误] 编译目录不存在: %s", srcDir)
	}
	fmt.Printf("[信息] 编译目录验证通过: %s\r\n", srcDir)

	// 定义要构建的平台列表
	buildTargets := []struct {
		os  string
		ext string
	}{
		{"windows", ".exe"},
		{"linux", ""},
	}
	fmt.Println("[信息] 构建平台已确定: Windows, Linux")

	// 获取应用名称
	appName := filepath.Base(cmdDir)
	fmt.Printf("[信息] 应用名称已获取: %s\r\n", appName)

	// 获取Git信息
	gitInfo, err := getGitInfo(rootDir)
	if err != nil {
		return fmt.Errorf("[错误] 获取Git信息失败: %w", err)
	}

	fmt.Println("\n[信息] 主仓库Git信息:")
	fmt.Println(gitInfo.String())
	mainGitInfo := gitInfo.String()

	if len(gitInfo.UncommittedFiles) > 0 {
		fs := ""
		for _, f := range gitInfo.UncommittedFiles {
			fs += fmt.Sprintf("   %s\r\n", f)
		}

		if forceBuild {
			fmt.Println(fmt.Sprintf("[警告] 主仓库存在以下未提交的文件:\r\n%s", fs))
			fmt.Println("[信息] 已启用强制编译,继续执行构建")
			mainGitInfo += "警告: 编译时存在以下未提交文件,强制编译模式\r\n"
			mainGitInfo += fs
		} else {
			return fmt.Errorf("[错误] 无法构建,主仓库存在以下未提交的文件,请先提交变更或使用 -f 强制编译\r\n%s", fs)
		}
	}

	// 新增: 获取并打印依赖库Git信息
	var depsInfo []DepGitInfo
	depsInfo, err = getDependenciesGitInfo(rootDir)
	if err != nil {
		return fmt.Errorf("[错误] 获取依赖库Git信息失败: %w", err)
	}

	fmt.Println("\n[信息] 依赖库Git信息:")
	depgitInfo := "依赖库信息：\r\n"

	for _, dep := range depsInfo {

		msg := fmt.Sprintf(
			"  依赖库: %s\r\n"+
				"    构建分支: %s\r\n"+
				"    提 交 ID: %s\r\n"+
				"    提交信息: %s\r\n"+
				"    提交用户: %s\r\n"+
				"    提交时间: %s\r\n",
			filepath.Base(dep.Path),
			dep.Git.Branch,
			dep.Git.CommitID,
			dep.Git.CommitMsg,
			dep.Git.CommitAuthor,
			dep.Git.CommitTime)
		fmt.Printf("\n[信息] 依赖库%s,Git信息:\r\n%s", filepath.Base(dep.Path), msg)
		depgitInfo += msg
		if len(dep.Git.UncommittedFiles) > 0 {
			fs := ""
			for _, f := range dep.Git.UncommittedFiles {
				fs += fmt.Sprintf("    %s\r\n", f)
			}
			if forceBuild {
				fmt.Println(fmt.Sprintf("[警告] 依赖库%s,存在以下未提交的文件:\r\n%s", filepath.Base(dep.Path), fs))
				fmt.Println("[信息] 已启用强制编译,继续执行构建")
				depgitInfo += "  警告:编译时存在以下未提交文件,强制编译模式\r\n"
				depgitInfo += fs
			} else {
				return fmt.Errorf("[错误] 无法构建: 依赖库 %s 存在未提交变更,请先提交或使用 -f 强制编译\r\n", filepath.Base(dep.Path))
			}
		}
	}

	// 遍历平台列表进行构建
	for _, target := range buildTargets {
		if err := buildForOS(rootDir, target.os, target.ext, appName, version, cmdDir, mainGitInfo, depgitInfo); err != nil {
			return fmt.Errorf("[错误] %s 平台构建失败: %w", target.os, err)
		}
	}

	elapsed := time.Since(startTime).Round(time.Millisecond)
	fmt.Printf("\n[成功] 所有平台构建完成! 总耗时: %s\r\n", elapsed)
	return nil
}

// buildForOS 为特定操作系统执行构建
func buildForOS(rootDir, goos, ext, appName, version, cmdDir string, mainGitInfo string, depGitInfo string) error {
	fmt.Printf("\n[信息] 开始构建 %s 平台...\n", goos)
	startTime := time.Now()

	// 设置环境变量
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOOS=%s", goos))
	env = append(env, "GOARCH=amd64") // 默认为amd64架构

	// 创建输出目录,包含版本号
	outputDir := filepath.Join(filepath.Dir(rootDir), "bin", "build_out_"+appName, version)
	outputDir = filepath.Clean(outputDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("[错误] 创建输出目录失败: 路径=%s, 错误=%w", outputDir, err)
	}
	fmt.Printf("[信息] 输出目录已创建: %s\r\n", outputDir)

	// 构建命令
	outputPath := filepath.Join(outputDir, appName+ext)
	cmd := exec.Command("go", "build",
		"-ldflags", fmt.Sprintf(
			"-X main.Model=release"+
				" -X main.Version=%s"+
				" -X main.MainGit=%s"+
				" -X main.DepGits=%s",
			version,
			FilerStr(mainGitInfo),
			FilerStr(depGitInfo),
		),
		"-o", outputPath,
		fmt.Sprintf("./cmd/%s", cmdDir))

	cmd.Dir = rootDir
	cmd.Env = env

	// 打印构建命令信息
	cmdStr := strings.Join(cmd.Args, " ")
	fmt.Printf("[信息] 执行构建命令: %s\r\n", cmdStr)
	fmt.Printf("[信息] 输出文件路径: %s\r\n", outputPath)

	// 执行命令并捕获输出
	fmt.Println("[信息] 开始执行构建...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[错误] 构建命令执行失败: %v\n", err)
		fmt.Printf("[错误详情] %s\r\n", string(output))
		return fmt.Errorf("执行构建命令失败: %w\n输出: %s", err, string(output))
	}

	// 输出构建成功信息
	elapsed := time.Since(startTime).Round(time.Millisecond)
	fmt.Printf("[成功] %s 平台构建完成! 输出文件: %s 耗时: %s\r\n", goos, outputPath, elapsed)
	return nil
}

// findGoProjectRootDir 通过查找go.mod文件来确定项目的根目录
func findGoProjectRootDir() (string, error) {
	fmt.Println("[信息] 开始查找项目根目录")

	// 获取当前工作目录
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("[错误] 获取当前目录失败: %w", err)
	}
	fmt.Printf("[信息] 当前工作目录已获取: %s\r\n", pwd)

	current := pwd
	for {
		// 检查当前目录下是否存在go.mod文件
		goModPath := filepath.Join(current, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			fmt.Printf("[信息] 项目根目录查找完成: %s\r\n", current)
			return current, nil
		}

		// 获取上一级目录
		next := filepath.Dir(current)

		// 如果已经到达根目录（即没有上一级目录）,则返回错误
		if next == current {
			return "", fmt.Errorf("[错误] 项目根目录查找失败: 在任何父目录中都找不到go.mod")
		}

		current = next
	}
}

// getAppName 从main包中提取应用名称
func getAppName(rootDir string) string {
	// 默认使用目录名
	dirName := filepath.Base(rootDir)
	fmt.Printf("[信息] 应用名称默认值已设置: %s\r\n", dirName)

	// 尝试从cmd目录推断应用名称
	cmdDir := filepath.Join(rootDir, "cmd")
	if info, err := os.Stat(cmdDir); err == nil && info.IsDir() {
		fmt.Println("[信息] 正在从cmd目录推断应用名称")
		entries, err := os.ReadDir(cmdDir)
		if err == nil && len(entries) > 0 {
			for _, entry := range entries {
				if entry.IsDir() {
					fmt.Printf("[信息] 应用名称已从cmd目录获取: %s\r\n", entry.Name())
					return entry.Name()
				}
			}
		}
	}

	fmt.Println("[信息] 未从cmd目录找到应用名称,使用默认值")
	return dirName
}

// GitInfo 包含从Git获取的版本控制信息
type GitInfo struct {
	Branch           string   // 当前分支
	CommitID         string   // 提交ID
	CommitMsg        string   // 提交信息
	CommitAuthor     string   // 提交用户
	CommitTime       string   // 提交时间
	BuildUser        string   // 构建用户
	BuildTime        string   // 构建时间
	UncommittedFiles []string // 未提交的文件列表
}

func (g *GitInfo) String() string {
	return fmt.Sprintf(
		"构建信息:\r\n"+
			"  构建分支:%s\r\n"+
			"  提 交 ID:%s\r\n"+
			"  提交信息:%s\r\n"+
			"  提交用户:%s\r\n"+
			"  提交时间:%s\r\n"+
			"  构建用户:%s\r\n"+
			"  构建时间:%s\r\n",
		g.Branch, g.CommitID, g.CommitMsg, g.CommitAuthor, g.CommitTime, g.BuildUser, g.BuildTime)
}

func FilerStr(in string) string {
	str := strings.ReplaceAll(in, " ", "&nbsp")
	str = strings.ReplaceAll(str, "\r\n", "<br>")
	return str
}

// getGitInfo 从Git获取版本控制信息
func getGitInfo(projectDir string) (GitInfo, error) {
	var info GitInfo
	fmt.Println("[信息] 开始获取主仓库Git信息")

	// 获取当前分支
	branch, err := execGitCommand(projectDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		branch = "unknown"
		fmt.Printf("[警告] 获取分支信息失败,使用默认值: %s\r\n", branch)
	}
	info.Branch = branch
	fmt.Printf("[信息] 分支信息已获取: %s\r\n", branch)

	// 获取提交ID
	commitID, err := execGitCommand(projectDir, "rev-parse", "--short", "HEAD")
	if err != nil {
		commitID = "unknown"
		fmt.Printf("[警告] 获取提交ID失败,使用默认值: %s\r\n", commitID)
	}
	info.CommitID = commitID
	fmt.Printf("[信息] 提交ID已获取: %s\r\n", commitID)

	// 获取提交信息
	commitMsg, err := execGitCommand(projectDir, "log", "-1", "--pretty=format:%s")
	if err != nil {
		commitMsg = "unknown"
		fmt.Printf("[警告] 获取提交信息失败,使用默认值: %s\r\n", commitMsg)
	}
	info.CommitMsg = commitMsg
	fmt.Printf("[信息] 提交信息已获取: %s\r\n", commitMsg)

	// 获取最后一次提交的作者（标准格式：姓名 <邮箱>）
	author, err := execGitCommand(projectDir, "log", "-1", "--pretty=format:%cN <%cE>")
	if err != nil {
		author = "unknown <unknown@example.com>"
		fmt.Printf("[警告] 获取提交用户失败,使用默认值: %s\r\n", author)
	}
	info.CommitAuthor = author
	fmt.Printf("[信息] 提交用户已获取: %s\r\n", author)

	// 获取最后一次提交的时间（优化时间格式为本地时间）
	commitTime, err := execGitCommand(projectDir, "log", "-1", "--pretty=format:%ci")
	if err != nil {
		commitTime = time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("[警告] 获取提交时间失败,使用当前时间: %s\r\n", commitTime)
	}
	info.CommitTime = commitTime
	fmt.Printf("[信息] 提交时间已获取: %s\r\n", commitTime)

	// 获取Git用户名
	user, err := execGitCommand(projectDir, "config", "user.name")
	if err != nil {
		user = "unknown"
		fmt.Printf("[警告] 获取构建用户失败,使用默认值: %s\r\n", user)
	}
	info.BuildUser = user
	fmt.Printf("[信息] 构建用户已获取: %s\r\n", user)

	// 获取当前时间作为构建时间
	info.BuildTime = time.Now().Format("2006-01-02 15:04:05.000") // UTC时间
	fmt.Printf("[信息] 构建时间已获取: %s\r\n", info.BuildTime)

	// 检测未提交的文件
	info.UncommittedFiles = checkUncommittedFiles(projectDir)
	fmt.Printf("[信息] 未提交文件检测完成: 共 %d 个未提交文件\n", len(info.UncommittedFiles))

	return info, nil
}

// execGitCommand 执行Git命令并返回输出
func execGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// checkUncommittedFiles 检测本地仓库中未提交的文件,返回带状态标记的文件数组
func checkUncommittedFiles(projectDir string) []string {
	var uncommittedFiles []string
	fmt.Println("[信息] 开始检测未提交文件")

	// 检查是否有未提交的变更
	isClean, err := execGitCommand(projectDir, "status", "--porcelain")
	if err != nil {
		fmt.Printf("[警告] 无法检测未提交变更: %v\n", err)
		return uncommittedFiles
	}

	// 如果输出非空,说明有未提交的变更
	if isClean != "" {
		fmt.Println("[信息] 检测到未提交变更,正在分析...")

		// 1. 获取已修改但未暂存的文件
		modifiedFiles, err := execGitCommand(projectDir, "diff", "--name-only")
		if err == nil {
			for _, file := range strings.Split(modifiedFiles, "\n") {
				if file != "" {
					uncommittedFiles = append(uncommittedFiles, "修改: "+file)
				}
			}
		} else {
			fmt.Printf("[警告] 获取已修改文件失败: %v\n", err)
		}

		// 2. 获取已暂存但未提交的文件
		stagedFiles, err := execGitCommand(projectDir, "diff", "--staged", "--name-only")
		if err == nil {
			for _, file := range strings.Split(stagedFiles, "\n") {
				if file != "" {
					uncommittedFiles = append(uncommittedFiles, "暂存: "+file)
				}
			}
		} else {
			fmt.Printf("[警告] 获取已暂存文件失败: %v\n", err)
		}

		// 3. 获取未跟踪的文件
		untrackedFiles, err := execGitCommand(projectDir, "ls-files", "--others", "--exclude-standard")
		if err == nil {
			for _, file := range strings.Split(untrackedFiles, "\n") {
				if file != "" {
					uncommittedFiles = append(uncommittedFiles, "新增: "+file)
				}
			}
		} else {
			fmt.Printf("[警告] 获取未跟踪文件失败: %v\n", err)
		}
	} else {
		fmt.Println("[信息] 未检测到未提交变更")
	}

	return uncommittedFiles
}

// 新增: 依赖库Git信息结构
type DepGitInfo struct {
	Path string  // 依赖库路径
	Git  GitInfo // 依赖库Git信息
}

// 新增: 获取依赖库Git信息（通过解析go.mod的replace指令）
func getDependenciesGitInfo(projectDir string) ([]DepGitInfo, error) {
	fmt.Println("[信息] 开始获取依赖库Git信息")

	// 读取go.mod文件
	goModPath := filepath.Join(projectDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("[错误] 读取go.mod失败: %w", err)
	}
	fmt.Printf("[信息] go.mod文件已读取: %s\r\n", goModPath)

	// 定义正则表达式匹配replace指令
	replaceRe := regexp.MustCompile(`(?m)^replace\s+([^\s]+)(\s+[^\s]+)?\s+=>\s+((\.\./|/|\w:\\|\\)[^\s]+)(\s+[^\s]+)?$`)
	lines := strings.Split(string(content), "\n")
	var localDeps []string
	var seenPaths = make(map[string]bool)
	fmt.Println("[信息] 开始解析replace指令")

	// 解析replace指令提取本地路径
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "replace") {
			continue
		}

		matches := replaceRe.FindStringSubmatch(line)
		if len(matches) < 4 {
			continue // 非有效replace行
		}

		// 提取替换路径
		rawPath := matches[3]

		// 处理Windows路径（支持反斜杠和驱动器号）
		cleanPath := filepath.Clean(strings.ReplaceAll(rawPath, "\\", "/"))

		// 处理相对路径（基于go.mod所在目录）
		var absPath string
		if filepath.IsAbs(cleanPath) {
			absPath = cleanPath
		} else {
			absPath, err = filepath.Abs(filepath.Join(projectDir, cleanPath))
			if err != nil {
				fmt.Printf("[警告] 解析路径失败: %s, 错误: %v\n", cleanPath, err)
				continue
			}
		}

		// 去重处理
		if !seenPaths[absPath] {
			seenPaths[absPath] = true
			localDeps = append(localDeps, absPath)
		}
	}
	fmt.Printf("[信息] replace指令解析完成: 找到 %d 个依赖库\n", len(localDeps))

	// 过滤出Git仓库并获取信息
	var depsInfo []DepGitInfo
	for i, path := range localDeps {
		fmt.Printf("\n[信息] 处理依赖库 #%d: %s\r\n", i+1, path)

		// 尝试查找.git目录
		gitDir, err := findGitDir(path)
		if err != nil {
			fmt.Printf("[警告] 未在 %s 及其父目录找到.git目录\n", path)
			continue
		}
		fmt.Printf("[信息] 依赖库Git目录已找到: %s\r\n", gitDir)

		// 获取Git信息（使用提交者信息而非本地配置）
		gitInfo, err := getGitInfoWithCommitter(gitDir)
		if err != nil {
			fmt.Printf("[警告] 获取 %s 的Git信息失败: %v\n", gitDir, err)
			continue
		}

		depsInfo = append(depsInfo, DepGitInfo{
			Path: path,
			Git:  gitInfo,
		})
	}

	return depsInfo, nil
}

// 获取Git信息,使用提交者信息而非本地Git配置
func getGitInfoWithCommitter(projectDir string) (GitInfo, error) {
	var info GitInfo

	// 获取当前分支
	branch, err := execGitCommand(projectDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		branch = "unknown"
	}
	info.Branch = branch

	// 获取提交ID
	commitID, err := execGitCommand(projectDir, "rev-parse", "--short", "HEAD")
	if err != nil {
		commitID = "unknown"
	}
	info.CommitID = commitID

	// 获取提交信息
	commitMsg, err := execGitCommand(projectDir, "log", "-1", "--pretty=format:%s")
	if err != nil {
		commitMsg = "unknown"
	}
	info.CommitMsg = commitMsg

	// 获取最后一次提交的作者（标准格式：姓名 <邮箱>）
	author, err := execGitCommand(projectDir, "log", "-1", "--pretty=format:%cN <%cE>")
	if err != nil {
		author = "unknown <unknown@example.com>"
	}
	info.CommitAuthor = author

	// 获取最后一次提交的时间（优化时间格式为本地时间）
	commitTime, err := execGitCommand(projectDir, "log", "-1", "--pretty=format:%ci")
	if err != nil {
		commitTime = time.Now().Format("2006-01-02 15:04:05")
	}
	info.CommitTime = commitTime

	// 检测未提交的文件
	info.UncommittedFiles = checkUncommittedFiles(projectDir)
	return info, nil
}

// 辅助函数：查找目录下的.git目录（增强版）
func findGitDir(startDir string) (string, error) {
	// 逐级向上查找.git目录
	currentDir := startDir
	maxDepth := 5
	depth := 0

	for {
		// 检查当前目录是否存在.git
		gitDir := filepath.Join(currentDir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return currentDir, nil
		}

		// 检查是否为Git工作树（外部.git目录）
		gitFile := filepath.Join(currentDir, ".git")
		if info, err := os.Stat(gitFile); err == nil && !info.IsDir() {
			// 解析.git文件中的gitdir路径
			content, err := os.ReadFile(gitFile)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "gitdir: ") {
						externalGitDir := strings.TrimPrefix(line, "gitdir: ")
						externalGitDir = strings.TrimSpace(externalGitDir)

						// 处理相对路径
						if !filepath.IsAbs(externalGitDir) {
							externalGitDir = filepath.Join(currentDir, externalGitDir)
						}

						if _, err := os.Stat(externalGitDir); err == nil {
							return filepath.Dir(externalGitDir), nil
						}
					}
				}
			}
		}

		// 向上移动一级
		parentDir := filepath.Dir(currentDir)

		// 检查是否到达根目录或超出最大深度
		if parentDir == currentDir || depth >= maxDepth {
			return "", fmt.Errorf("未找到.git目录")
		}

		currentDir = parentDir
		depth++
	}
}
