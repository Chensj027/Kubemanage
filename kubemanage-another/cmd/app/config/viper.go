package config

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var configObj = &Config{
	Default: DefaultOptions{
		ListenAddr:          ":6180",
		WebSocketListenAddr: "",
		JWTSecret:           "kubemanage",
		ExpireTime:          10,
	},
	Mysql: MysqlOptions{
		Host:         "127.0.0.1",
		Port:         "3306",
		User:         "root",
		Password:     "change-me",
		Name:         "kubemanage",
		MaxOpenConns: 100,
		MaxLifetime:  20,
		MaxIdleConns: 10,
	},
	Log: LogConfig{
		Level:      "debug",
		Filename:   "kubemanage.log",
		MaxSize:    200,
		MaxAge:     30,
		MaxBackups: 7,
	},
}

func defaultConfig() *Config {
	cfg := *configObj
	return &cfg
}

// Binding 解析外部的配置文件，默认是 ./config.yaml
func Binding(filePath string) error {
	v := viper.New()
	SysConfig = defaultConfig()

	if filePath == "" {
		return nil
	}

	v.SetConfigFile(filePath)
	if err := v.ReadInConfig(); err != nil {
		if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
			return nil
		}
		return fmt.Errorf("read config file %q failed: %w", filePath, err)
	}

	if err := v.Unmarshal(SysConfig); err != nil {
		return fmt.Errorf("config unmarshal failed: %w", err)
	}

	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("config file changed,sys config reload")
		if err := v.Unmarshal(SysConfig); err != nil {
			fmt.Printf("config file changed,viper.Unmarshal failed, err:%v\n", err)
		}
	})
	return nil
}
