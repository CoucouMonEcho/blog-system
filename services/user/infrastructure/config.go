package infrastructure

import (
	"blog-system/common/pkg/logger"
	"os"

	"gopkg.in/yaml.v2"
)

// RedisClusterConfig Redis集群配置
type RedisClusterConfig struct {
	Addrs        []string `yaml:"addrs"`
	Password     string   `yaml:"password"`
	PoolSize     int      `yaml:"pool_size"`
	MinIdleConns int      `yaml:"min_idle_conns"`
	MaxRetries   int      `yaml:"max_retries"`
	DialTimeout  string   `yaml:"dial_timeout"`
	ReadTimeout  string   `yaml:"read_timeout"`
	WriteTimeout string   `yaml:"write_timeout"`
}

// AppConfig 映射 user.yaml
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
		// Redis Cluster配置
		Cluster RedisClusterConfig `yaml:"cluster"`
	} `yaml:"redis"`
	Grpc struct {
		Port int `yaml:"port"`
	} `yaml:"grpc"`
	Log logger.Config `yaml:"log"`
}

// LoadConfig 读取配置
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
