package infrastructure

import (
	"blog-system/common/pkg/logger"
	"os"

	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	App struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	} `yaml:"app"`
	Database struct {
		Driver   string `yaml:"driver"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`
	Redis struct {
		Cluster struct {
			Addrs        []string `yaml:"addrs"`
			Password     string   `yaml:"password"`
			PoolSize     int      `yaml:"pool_size"`
			MinIdleConns int      `yaml:"min_idle_conns"`
			MaxRetries   int      `yaml:"max_retries"`
			DialTimeout  string   `yaml:"dial_timeout"`
			ReadTimeout  string   `yaml:"read_timeout"`
			WriteTimeout string   `yaml:"write_timeout"`
		} `yaml:"cluster"`
	} `yaml:"redis"`
	Registry struct {
		Endpoints []string `yaml:"endpoints"`
	} `yaml:"registry"`
	UserService struct {
		// service://user-service （通过注册中心解析）或 http://host:port
		BaseURL string `yaml:"base_url"`
		// 请求超时毫秒
		Timeout int `yaml:"timeout"`
	} `yaml:"user_service"`
	Log logger.Config `yaml:"log"`
}

func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
