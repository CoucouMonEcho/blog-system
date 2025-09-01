package config

import (
	"os"

	"blog-system/common/pkg/logger"

	"gopkg.in/yaml.v2"
)

// RedisClusterConfig defines redis cluster settings
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

// RouteConfig Gateway specific route config (optional in unified config)
type RouteConfig struct {
	User    Route `yaml:"user"`
	Content Route `yaml:"content"`
	Stat    Route `yaml:"stat"`
	Admin   Route `yaml:"admin"`
}

type Route struct {
	Prefix  string `yaml:"prefix"`
	Target  string `yaml:"target"`
	Timeout string `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}

type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerSecond int  `yaml:"requests_per_second"`
	Burst             int  `yaml:"burst"`
}

type CircuitBreakerConfig struct {
	Enabled          bool   `yaml:"enabled"`
	FailureThreshold int    `yaml:"failure_threshold"`
	RecoveryTimeout  string `yaml:"recovery_timeout"`
}

// AppConfig is the unified configuration for all services
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
		Cluster RedisClusterConfig `yaml:"cluster"`
	} `yaml:"redis"`
	Registry struct {
		Endpoints []string `yaml:"endpoints"`
	} `yaml:"registry"`
	GRPC struct {
		Port int `yaml:"port"`
	} `yaml:"grpc"`
	Routes         RouteConfig          `yaml:"routes"`
	RateLimit      RateLimitConfig      `yaml:"rate_limit"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	Prometheus     struct {
		Address string `yaml:"address"`
	} `yaml:"prometheus"`
	Log logger.Config `yaml:"log"`
}

// ResolvePath tries typical locations for service config
func ResolvePath(service string) string {
	// prefer absolute path in deployment
	if _, err := os.Stat("/opt/blog-system/configs/" + service + ".yaml"); err == nil {
		return "/opt/blog-system/configs/" + service + ".yaml"
	}
	// development path from service module
	if _, err := os.Stat("../../configs/" + service + ".yaml"); err == nil {
		return "../../configs/" + service + ".yaml"
	}
	// fallback to repo root execution
	return "configs/" + service + ".yaml"
}

// Load loads config by service name using ResolvePath
func Load(service string) (*AppConfig, error) {
	return LoadByPath(ResolvePath(service))
}

// LoadByPath loads config from a specific path
func LoadByPath(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	applyEnvOverrides(&cfg)
	return &cfg, nil
}

// applyEnvOverrides allows password fields to reference env var names
func applyEnvOverrides(cfg *AppConfig) {
	if cfg == nil {
		return
	}
	if cfg.Database.Password != "" {
		if envVal := os.Getenv(cfg.Database.Password); envVal != "" {
			cfg.Database.Password = envVal
		}
	}
	if cfg.Redis.Cluster.Password != "" {
		if envVal := os.Getenv(cfg.Redis.Cluster.Password); envVal != "" {
			cfg.Redis.Cluster.Password = envVal
		}
	}
}
