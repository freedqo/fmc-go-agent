package ucsvvp

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gocarina/gocsv"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type CsvVp struct {
	vp      *viper.Viper
	csvData interface{}
	path    string
}

// New 初始化配置文件（默认配置文件目录：./config）
// param config 配置文件结构体（指针）
// param serverName 服务名称(日志文件名称)
// param log 日志函数
// 返回：错误
func New(csvData interface{}, serverName string) (*CsvVp, error) {
	if csvData == nil {
		return nil, errors.New("config struct is nil")
	}
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	//实例viper
	vp := viper.New()

	//读取config.yaml，并且部署热更新
	s := &CsvVp{vp: vp, csvData: csvData}

	configfileName := serverName + ".csv"
	SysConfigDir := filepath.Join(root, "csv")
	s.path = filepath.Join(SysConfigDir, configfileName)

	// 检测路径是否存在
	_, err = os.Stat(SysConfigDir)
	if err != nil {
		err = os.MkdirAll(SysConfigDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	//检查文件是否存在
	_, err2 := os.Stat(s.path)
	if err2 != nil {
		err3 := s.Write()
		if err3 != nil {
			return nil, err3
		}
	}

	// 设置配置文件名，配置文件路径，配置文件格式
	vp.SetConfigName(serverName)
	vp.AddConfigPath(SysConfigDir)
	vp.SetConfigType("csv")
	viper.SupportedExts = append(viper.SupportedExts, "csv")
	err = vp.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = s.Read(csvData)
	if err != nil {
		return nil, err
	}

	//文件写入本地
	err = s.Write()
	if err != nil {
		return nil, err
	}

	s.watchSettingChange(csvData)

	return s, nil
}

// watchSettingChange 监听文件变化，配置热更新
func (s *CsvVp) watchSettingChange(config interface{}) {
	go func() {
		s.vp.WatchConfig()
		s.vp.OnConfigChange(func(in fsnotify.Event) {
			err := s.Read(config)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("配置文件:%s,已更新", in.Name)
		})
	}()
}

func (s *CsvVp) Read(config interface{}) error {
	// 打开 CSV 文件
	file, err := os.OpenFile(s.path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("读取 CSV 失败: %v", err)
	}
	defer file.Close()
	// 从 CSV 文件解析到结构体切片
	if err := gocsv.UnmarshalFile(file, s.csvData); err != nil {
		return fmt.Errorf("解析 CSV 失败: %v", err)
	}
	return nil
}
func (s *CsvVp) Write() error {
	// 创建CSV文件
	file, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer file.Close()
	// 将结构体数组写入CSV
	if err = gocsv.MarshalFile(s.csvData, file); err != nil {
		return err
	}
	return nil

}
