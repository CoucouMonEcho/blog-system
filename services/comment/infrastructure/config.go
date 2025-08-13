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
	Registry struct {
		Endpoints []string `yaml:"endpoints"`
	} `yaml:"registry"`
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
	if cfg.Database.Password != "" {
		if envVal := os.Getenv(cfg.Database.Password); envVal != "" {
			cfg.Database.Password = envVal
		}
	}
	return &cfg, nil
}
