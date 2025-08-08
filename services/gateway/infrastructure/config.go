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

// RouteConfig 路由配置
type RouteConfig struct {
	User    Route `yaml:"user"`
	Content Route `yaml:"content"`
	Comment Route `yaml:"comment"`
	Stat    Route `yaml:"stat"`
	Admin   Route `yaml:"admin"`
}

// Route 路由规则
type Route struct {
	Prefix  string `yaml:"prefix"`
	Target  string `yaml:"target"`
	Timeout string `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerSecond int  `yaml:"requests_per_second"`
	Burst             int  `yaml:"burst"`
}

// CircuitBreakerConfig 熔断配置
type CircuitBreakerConfig struct {
	Enabled          bool   `yaml:"enabled"`
	FailureThreshold int    `yaml:"failure_threshold"`
	RecoveryTimeout  string `yaml:"recovery_timeout"`
}

// GatewayConfig 网关配置
type GatewayConfig struct {
	App struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	} `yaml:"app"`
	Routes         RouteConfig          `yaml:"routes"`
	RateLimit      RateLimitConfig      `yaml:"rate_limit"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	Redis          struct {
		Cluster RedisClusterConfig `yaml:"cluster"`
	} `yaml:"redis"`
	Log logger.Config `yaml:"log"`
}

// LoadConfig 读取配置
func LoadConfig(path string) (*GatewayConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg GatewayConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
