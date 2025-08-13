package application

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"blog-system/common/pkg/logger"
	"blog-system/services/gateway/domain"
)

// GatewayService 网关服务
type GatewayService struct {
	routeRepo        domain.RouteRepository
	serviceDiscovery domain.ServiceDiscovery
	rateLimiter      domain.RateLimiter
	circuitBreaker   domain.CircuitBreaker
	logger           logger.Logger
}

// NewGatewayService 创建网关服务
func NewGatewayService(
	routeRepo domain.RouteRepository,
	serviceDiscovery domain.ServiceDiscovery,
	rateLimiter domain.RateLimiter,
	circuitBreaker domain.CircuitBreaker,
	lgr logger.Logger,
) *GatewayService {
	return &GatewayService{
		routeRepo:        routeRepo,
		serviceDiscovery: serviceDiscovery,
		rateLimiter:      rateLimiter,
		circuitBreaker:   circuitBreaker,
		logger:           lgr,
	}
}

// ProxyRequest 代理请求到目标服务
func (s *GatewayService) ProxyRequest(ctx context.Context, req *domain.ProxyRequest) (*domain.ProxyResponse, error) {
	// 1. 限流检查
	if s.rateLimiter != nil && !s.rateLimiter.Allow(req.Client) {
		s.logger.LogWithContext("gateway-service", "application", "WARN", "限流触发: client=%s path=%s", req.Client, req.Path)
		return &domain.ProxyResponse{
			StatusCode: http.StatusTooManyRequests,
			Body:       []byte("请求过于频繁，请稍后再试"),
		}, nil
	}

	// 2. 路由匹配
	route := s.routeRepo.GetRouteByPath(req.Path)
	if route == nil {
		s.logger.LogWithContext("gateway-service", "application", "WARN", "路由未命中: path=%s", req.Path)
		return &domain.ProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       []byte("路由不存在"),
		}, nil
	}

	// 3. 熔断检查
	if s.circuitBreaker != nil && s.circuitBreaker.IsOpen(route.Target) {
		s.logger.LogWithContext("gateway-service", "application", "WARN", "熔断开启: target=%s", route.Target)
		return &domain.ProxyResponse{
			StatusCode: http.StatusServiceUnavailable,
			Body:       []byte("服务暂时不可用"),
		}, nil
	}

	// 4. 解析 service:// 目标
	targetStr := route.Target
	if s.serviceDiscovery != nil {
		if resolved, ok := s.serviceDiscovery.Resolve(route.Target); ok {
			targetStr = resolved
		}
	}

	// 5. 服务健康检查
	if s.serviceDiscovery != nil && !s.serviceDiscovery.GetServiceHealth(targetStr) {
		if s.circuitBreaker != nil {
			s.circuitBreaker.RecordFailure(targetStr)
		}
		s.logger.LogWithContext("gateway-service", "application", "WARN", "目标不健康: target=%s", targetStr)
		return &domain.ProxyResponse{
			StatusCode: http.StatusServiceUnavailable,
			Body:       []byte("目标服务不可用"),
		}, nil
	}

	// 6. 构建目标URL
	targetURL, err := url.Parse(targetStr)
	if err != nil {
		s.logger.LogWithContext("gateway-service", "application", "ERROR", "解析目标失败: target=%s err=%v", targetStr, err)
		return &domain.ProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       []byte("路由配置错误"),
		}, err
	}

	// 7. 计算转发路径：/api/{service}/x -> /api/x
	forwardPath := req.Path
	if strings.HasPrefix(route.Prefix, "/api/") && strings.HasPrefix(req.Path, route.Prefix) {
		forwardPath = "/api" + strings.TrimPrefix(req.Path, route.Prefix)
	}

	// 8. 创建HTTP客户端
	client := &http.Client{
		Timeout: route.Timeout,
	}

	// 9. 构建请求（使用字节Reader，避免编码问题）
	proxyReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL.String()+forwardPath, bytes.NewReader(req.Body))
	if err != nil {
		s.logger.LogWithContext("gateway-service", "application", "ERROR", "构建请求失败: url=%s err=%v", targetURL.String()+forwardPath, err)
		return &domain.ProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       []byte("创建请求失败"),
		}, err
	}

	// 10. 复制端到端请求头，过滤 hop-by-hop 头
	hopByHop := map[string]struct{}{
		"Connection":          {},
		"Proxy-Connection":    {},
		"Keep-Alive":          {},
		"Proxy-Authenticate":  {},
		"Proxy-Authorization": {},
		"Te":                  {},
		"Trailer":             {},
		"Transfer-Encoding":   {},
		"Upgrade":             {},
		"Content-Length":      {},
		"Expect":              {},
	}
	for key, values := range req.Headers {
		ck := http.CanonicalHeaderKey(key)
		if _, skip := hopByHop[ck]; skip {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(ck, value)
		}
	}
	// 追加 X-Forwarded-For
	if prior, ok := proxyReq.Header["X-Forwarded-For"]; ok && len(prior) > 0 {
		proxyReq.Header.Set("X-Forwarded-For", prior[0]+", "+req.Client)
	} else {
		proxyReq.Header.Set("X-Forwarded-For", req.Client)
	}

	// 11. 发送请求
	resp, err := client.Do(proxyReq)
	if err != nil {
		if s.circuitBreaker != nil {
			s.circuitBreaker.RecordFailure(targetStr)
		}
		s.logger.LogWithContext("gateway-service", "application", "ERROR", "请求目标失败: target=%s err=%v", targetStr, err)
		return &domain.ProxyResponse{
			StatusCode: http.StatusBadGateway,
			Body:       []byte("请求目标服务失败"),
		}, err
	}
	defer func() { _ = resp.Body.Close() }()

	// 12. 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.LogWithContext("gateway-service", "application", "ERROR", "读取响应失败: target=%s err=%v", targetStr, err)
		return &domain.ProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       []byte("读取响应失败"),
		}, err
	}

	// 13. 记录成功
	if s.circuitBreaker != nil {
		s.circuitBreaker.RecordSuccess(targetStr)
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
