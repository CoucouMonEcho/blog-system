package application

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"blog-system/services/gateway/domain"
)

// GatewayService 网关服务
type GatewayService struct {
	routeRepo        domain.RouteRepository
	serviceDiscovery domain.ServiceDiscovery
	rateLimiter      domain.RateLimiter
	circuitBreaker   domain.CircuitBreaker
}

// NewGatewayService 创建网关服务
func NewGatewayService(
	routeRepo domain.RouteRepository,
	serviceDiscovery domain.ServiceDiscovery,
	rateLimiter domain.RateLimiter,
	circuitBreaker domain.CircuitBreaker,
) *GatewayService {
	return &GatewayService{
		routeRepo:        routeRepo,
		serviceDiscovery: serviceDiscovery,
		rateLimiter:      rateLimiter,
		circuitBreaker:   circuitBreaker,
	}
}

// ProxyRequest 代理请求到目标服务
func (s *GatewayService) ProxyRequest(ctx context.Context, req *domain.ProxyRequest) (*domain.ProxyResponse, error) {
	// 1. 限流检查
	if s.rateLimiter != nil && !s.rateLimiter.Allow(req.Client) {
		return &domain.ProxyResponse{
			StatusCode: http.StatusTooManyRequests,
			Body:       []byte("请求过于频繁，请稍后再试"),
		}, nil
	}

	// 2. 路由匹配
	route := s.routeRepo.GetRouteByPath(req.Path)
	if route == nil {
		return &domain.ProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       []byte("路由不存在"),
		}, nil
	}

	// 3. 熔断检查
	if s.circuitBreaker != nil && s.circuitBreaker.IsOpen(route.Target) {
		return &domain.ProxyResponse{
			StatusCode: http.StatusServiceUnavailable,
			Body:       []byte("服务暂时不可用"),
		}, nil
	}

	// 4. 服务健康检查
	if s.serviceDiscovery != nil && !s.serviceDiscovery.GetServiceHealth(route.Target) {
		if s.circuitBreaker != nil {
			s.circuitBreaker.RecordFailure(route.Target)
		}
		return &domain.ProxyResponse{
			StatusCode: http.StatusServiceUnavailable,
			Body:       []byte("目标服务不可用"),
		}, nil
	}

	// 5. 构建目标URL
	targetURL, err := url.Parse(route.Target)
	if err != nil {
		return &domain.ProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       []byte("路由配置错误"),
		}, err
	}

	// 6. 创建HTTP客户端
	client := &http.Client{
		Timeout: route.Timeout,
	}

	// 7. 构建请求
	proxyReq, err := http.NewRequest(req.Method, targetURL.String()+req.Path, strings.NewReader(string(req.Body)))
	if err != nil {
		return &domain.ProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       []byte("创建请求失败"),
		}, err
	}

	// 8. 复制请求头
	for key, values := range req.Headers {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// 9. 发送请求
	resp, err := client.Do(proxyReq)
	if err != nil {
		if s.circuitBreaker != nil {
			s.circuitBreaker.RecordFailure(route.Target)
		}
		return &domain.ProxyResponse{
			StatusCode: http.StatusBadGateway,
			Body:       []byte("请求目标服务失败"),
		}, err
	}
	defer resp.Body.Close()

	// 10. 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &domain.ProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       []byte("读取响应失败"),
		}, err
	}

	// 11. 记录成功
	if s.circuitBreaker != nil {
		s.circuitBreaker.RecordSuccess(route.Target)
	}

	return &domain.ProxyResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}, nil
}

// GetRouteInfo 获取路由信息
func (s *GatewayService) GetRouteInfo(path string) *domain.Route {
	return s.routeRepo.GetRouteByPath(path)
}

// GetServiceHealth 获取服务健康状态
func (s *GatewayService) GetServiceHealth(target string) bool {
	if s.serviceDiscovery == nil {
		return true
	}
	return s.serviceDiscovery.GetServiceHealth(target)
}

// GetServiceLatency 获取服务延迟
func (s *GatewayService) GetServiceLatency(target string) time.Duration {
	if s.serviceDiscovery == nil {
		return 0
	}
	return s.serviceDiscovery.GetServiceLatency(target)
}
