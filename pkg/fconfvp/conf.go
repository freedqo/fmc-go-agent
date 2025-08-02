package fconfvp

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	encoderyaml "github.com/zwgblue/yaml-encoder"
	"os"
	"path/filepath"
)

type ConfigVp struct {
	vp     *viper.Viper
	config interface{}
	path   string
}

// New 初始化配置文件（默认配置文件目录：./config）
// param config 配置文件结构体（指针）
// param serverName 服务名称(日志文件名称)
// param log 日志函数
// 返回：错误
func New(config interface{}, serverName string) (*ConfigVp, error) {
	if config == nil {
		return nil, errors.New("config struct is nil")
	}
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	//实例viper
	vp := viper.New()

	//读取config.yaml，并且部署热更新
	s := &ConfigVp{vp: vp, config: config}

	configfileName := serverName + ".yaml"
	SysConfigDir := filepath.Join(root, "config")
	s.path = filepath.Join(SysConfigDir, configfileName)

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
	vp.SetConfigType("yaml")
	err = vp.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = s.reloadAllConfig(config)
	if err != nil {
		return nil, err
	}

	//文件写入本地
	err = s.Write()
	if err != nil {
		return nil, err
	}

	s.watchSettingChange(config)

	return s, nil
}

// watchSettingChange 监听文件变化，配置热更新
func (s *ConfigVp) watchSettingChange(config interface{}) {
	go func() {
		s.vp.WatchConfig()
		s.vp.OnConfigChange(func(in fsnotify.Event) {
			_ = s.reloadAllConfig(config)
			data, _ := json.Marshal(config)
			fmt.Printf("新配置文件路径：%s,内容:%v", in.Name, string(data))
		})
	}()
}

func (s *ConfigVp) reloadAllConfig(config interface{}) error {
	err := s.vp.Unmarshal(config)
	if err != nil {
		fmt.Printf("reloadAllConfig error!")
		return err
	}
	return nil
}

//	func (s *ConfigVp) Write() error {
//		encoder := encoderyaml.NewEncoder(s.config, encoderyaml.WithComments(encoderyaml.CommentsInLine))
//		content, _ := encoder.Encode()
//		var buf = bytes.Buffer{}
//		buf.Write(content)
//		return os.WriteFile(s.path, buf.Bytes(), 0777)
//	}
func (s *ConfigVp) Write() error {
	// 编码配置内容
	encoder := encoderyaml.NewEncoder(s.config, encoderyaml.WithComments(encoderyaml.CommentsInLine))
	content, err := encoder.Encode()
	if err != nil {
		return fmt.Errorf("编码配置内容失败: %w", err)
	}

	// 确保文件所在目录
	dirPath := filepath.Dir(s.path)

	// 创建目录目录不存在则创建，包括所有必要的父目录
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 写入文件，使用更合理的权限0644
	// 0644表示所有者可读写，其他用户只读
	if err := os.WriteFile(s.path, content, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}
