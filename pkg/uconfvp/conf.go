package uconfvp

import (
	"bytes"
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
func (s *ConfigVp) Write() error {
	// 创建目录
	dir := filepath.Dir(s.path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}

	encoder := encoderyaml.NewEncoder(s.config, encoderyaml.WithComments(encoderyaml.CommentsInLine))
	content, _ := encoder.Encode()
	var buf = bytes.Buffer{}
	buf.Write(content)
	return os.WriteFile(s.path, buf.Bytes(), 0777)
}
