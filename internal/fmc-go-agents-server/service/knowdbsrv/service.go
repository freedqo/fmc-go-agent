package knowdbsrv

import (
	"fmt"
	"github.com/freedqo/fmc-go-agents/internal/fmc-go-agents-server/model/knowdbm"
	"github.com/freedqo/fmc-go-agents/pkg/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func New() If {
	dir, err := os.Getwd()
	dir = filepath.Join(dir, "knowdb")
	if err != nil {
		panic(err)
	}
	s := &Service{
		docsDir: dir,
		files:   make(map[string]*knowdbm.TFileInfo),
	}
	if _, err := s.GetFileList(knowdbm.GetFileListReq{}); err != nil {
		panic(fmt.Sprintf("初始化文件列表失败: %v", err))
	}
	return s
}

type Service struct {
	docsDir string
	files   map[string]*knowdbm.TFileInfo
}

var _ If = &Service{}

func (s *Service) GetFileList(req knowdbm.GetFileListReq) (*knowdbm.GetFileListResp, error) {
	// 确保目录存在，不存在则创建
	if err := s.ensureUploadDirExists(); err != nil {
		return nil, fmt.Errorf("初始化文件目录失败: %v", err)
	}

	s.files = make(map[string]*knowdbm.TFileInfo) // 重置文件缓存
	var fileList []*knowdbm.TFileInfo

	// 遍历目录中的所有文件
	err := filepath.WalkDir(s.docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(s.docsDir, path)
		if err != nil {
			return err
		}

		md5Hash, err := utils.CalculateFileMD5(path)
		if err != nil {
			return fmt.Errorf("计算文件 %s 的MD5失败: %v", path, err)
		}

		// 获取文件类型（扩展名）
		ext := filepath.Ext(d.Name())
		fileType := strings.TrimPrefix(ext, ".")
		if fileType == "" {
			fileType = "unknown"
		}

		f := &knowdbm.TFileInfo{
			Id:   md5Hash,
			Name: d.Name(),
			Type: fileType,
			Path: relPath,
			Size: utils.FormatFileSize(fileInfo.Size()),
			Date: fileInfo.ModTime().Format(time.RFC3339),
		}
		fileList = append(fileList, f)
		s.files[f.Id] = f
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 按文件类型分组并排序
	dataFiles := make(map[string][]*knowdbm.TFileInfo)
	for _, file := range fileList {
		dataFiles[file.Type] = append(dataFiles[file.Type], file)
	}

	data := make([]*knowdbm.TypeList, 0, len(dataFiles))
	for k, v := range dataFiles {
		sort.Slice(v, func(i, j int) bool {
			return v[i].Date > v[j].Date // 按日期降序排列
		})
		data = append(data, &knowdbm.TypeList{
			Type:     k,
			FileList: v,
		})
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Type < data[j].Type
	})

	return &knowdbm.GetFileListResp{List: data}, nil
}

func (s *Service) DeleteFiles(req knowdbm.DeleteFilesReq) (res *knowdbm.DeleteFilesResp, err error) {
	for _, id := range req.Ids {
		file, ok := s.files[id]
		if !ok {
			return nil, fmt.Errorf("文件不存在，ID: %s", id)
		}

		filePath := filepath.Join(s.docsDir, file.Path)
		if err := os.Remove(filePath); err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("警告：文件 %s 不存在\n", filePath)
			} else {
				return nil, fmt.Errorf("删除文件失败: %v", err)
			}
		}
		delete(s.files, id)
	}

	return &knowdbm.DeleteFilesResp{}, nil
}

func (s *Service) Download(id string) (string, error) {
	file, ok := s.files[id]
	if !ok {
		return "", fmt.Errorf("文件不存在，ID: %s", id)
	}
	filePath := filepath.Join(s.docsDir, file.Path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", file)
	}
	return filePath, nil
}

func (s *Service) UploadFiles(files []*multipart.FileHeader) error {
	if err := s.ensureUploadDirExists(); err != nil {
		return err
	}
	for _, file := range files {
		// 1. 打开上传的文件，读取内容并计算MD5
		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("打开文件 %s 失败: %v", file.Filename, err)
		}

		// 读取文件内容（用于计算MD5和后续写入）
		content, err := io.ReadAll(src)
		src.Close() // 读取完成后立即关闭源文件，释放资源
		if err != nil {
			return fmt.Errorf("读取文件 %s 内容失败: %v", file.Filename, err)
		}
		// 2. 计算文件内容的MD5（判断是否已存在）
		md5Hash := utils.CalculateMD5(content) // 需确保utils有此方法：计算字节数组的MD5
		if fileInfo, exists := s.files[md5Hash]; exists {
			return fmt.Errorf("文件已存在（内容重复）,文件名称: %s", fileInfo.Name)
		}
		// 3. 处理文件类型目录
		ext := filepath.Ext(file.Filename)
		fileType := strings.TrimPrefix(ext, ".")
		if fileType == "" {
			fileType = "unknown" // 无扩展名文件归类到unknown目录
		}
		typeDir := filepath.Join(s.docsDir, fileType)
		if err := os.MkdirAll(typeDir, 0755); err != nil {
			return fmt.Errorf("创建类型目录 %s 失败: %v", typeDir, err)
		}
		// 4. 生成唯一文件名（避免同类型目录下文件名冲突）
		baseName := strings.TrimSuffix(file.Filename, ext)
		// 使用时间戳+原文件名前缀确保唯一性（即使MD5不同，文件名也可能重复）
		uniqueName := fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)
		dstPath := filepath.Join(typeDir, uniqueName)

		// 5. 写入文件到目标路径
		if err := os.WriteFile(dstPath, content, 0644); err != nil { // 0644：所有者可读写，其他只读
			return fmt.Errorf("保存文件到 %s 失败: %v", dstPath, err)
		}
	}
	// 刷新文件列表缓存
	if _, err := s.GetFileList(knowdbm.GetFileListReq{}); err != nil {
		return fmt.Errorf("刷新文件列表缓存失败: %v", err)
	}

	return nil
}

func (s *Service) ensureUploadDirExists() error {
	if _, err := os.Stat(s.docsDir); os.IsNotExist(err) {
		return os.MkdirAll(s.docsDir, 0755)
	}
	return nil
}
