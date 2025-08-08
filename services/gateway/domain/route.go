package domain

import (
	"net/http"
	"time"
)

// Route 路由规则
type Route struct {
	Prefix  string        `yaml:"prefix"`
	Target  string        `yaml:"target"`
	Timeout time.Duration `yaml:"timeout"`
	Retries int           `yaml:"retries"`
}

// RouteConfig 路由配置
type RouteConfig struct {
	User    Route `yaml:"user"`
	Content Route `yaml:"content"`
	Comment Route `yaml:"comment"`
	Stat    Route `yaml:"stat"`
	Admin   Route `yaml:"admin"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerSecond int  `yaml:"requests_per_second"`
	Burst             int  `yaml:"burst"`
}

// CircuitBreakerConfig 熔断配置
type CircuitBreakerConfig struct {
	Enabled          bool          `yaml:"enabled"`
	FailureThreshold int           `yaml:"failure_threshold"`
	RecoveryTimeout  time.Duration `yaml:"recovery_timeout"`
}

// ProxyRequest 代理请求
type ProxyRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
	Client  string // 客户端IP
}

// ProxyResponse 代理响应
type ProxyResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Error      error
}

// RouteRepository 路由仓储接口
type RouteRepository interface {
	GetRoutes() RouteConfig
	GetRouteByPath(path string) *Route
}

// ServiceDiscovery 服务发现接口
type ServiceDiscovery interface {
	GetServiceHealth(target string) bool
	GetServiceLatency(target string) time.Duration
}

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(client string) bool
	Reset(client string)
}

// CircuitBreaker 熔断器接口
type CircuitBreaker interface {
	IsOpen(target string) bool
	RecordSuccess(target string)
	RecordFailure(target string)
	TryReset(target string)
}
