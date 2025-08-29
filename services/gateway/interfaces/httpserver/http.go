package httpserver

import (
	"io"
	"net/http"
	"strconv"

	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/errcode"
	"blog-system/common/pkg/logger"
	"blog-system/services/gateway/application"
	"blog-system/services/gateway/domain"

	"github.com/CoucouMonEcho/go-framework/web"
	"github.com/CoucouMonEcho/go-framework/web/middlewares/accesslog"
	"github.com/CoucouMonEcho/go-framework/web/middlewares/errhandle"
	webprom "github.com/CoucouMonEcho/go-framework/web/middlewares/prometheus"
)

// HTTPServer HTTP服务器
type HTTPServer struct {
	gatewayService *application.GatewayService
	server         *web.HTTPServer
}

// NewHTTPServer 创建HTTP服务器
func NewHTTPServer(gatewayService *application.GatewayService) *HTTPServer {
	server := web.NewHTTPServer(
		web.ServerWithLogger(logger.Log().Error),
		web.ServerWithMiddlewares(
			errhandle.NewMiddlewareBuilder().RegisterError(http.StatusInternalServerError, []byte("内部服务错误")).Build(),
			accesslog.NewMiddlewareBuilder().LogFunc(func(log string) { logger.Log().Info(log) }).Build(),
			webprom.MiddlewareBuilder{Namespace: "blog-system", Subsystem: "gateway", Name: "http", Help: "gateway http latency"}.Build(),
		),
	)

	hs := &HTTPServer{
		gatewayService: gatewayService,
		server:         server,
	}

	// 挂载统一JWT鉴权（跳过健康检查与登录）
	server.Use(http.MethodGet, "/api/*", hs.authMiddleware())
	server.Use(http.MethodPost, "/api/*", hs.authMiddleware())

	return hs
}

// authMiddleware 统一JWT鉴权
func (s *HTTPServer) authMiddleware() web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			path := ctx.Req.URL.Path
			if path == "/health" || path == "/api/user/login" {
				next(ctx)
				return
			}
			authorization := ctx.Req.Header.Get("Authorization")
			uid, err := s.gatewayService.Authenticate(ctx.Req.Context(), authorization)
			if err != nil {
				logger.Log().Warn("httpserver: 鉴权失败: %v", err)
				_ = ctx.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrTokenInvalid, err.Error()))
				return
			}
			// 透传 X-User-ID，供后端使用
			ctx.Req.Header.Set("X-User-ID", strconv.FormatInt(uid, 10))
			next(ctx)
		}
	}
}

// Run 启动HTTP服务器
func (s *HTTPServer) Run(addr string) error {
	// 注册路由
	s.server.Get("/health", healthHandler())
	s.server.Post("/api/*", proxyHandler(s.gatewayService))
	s.server.Get("/api/*", proxyHandler(s.gatewayService))

	return s.server.Start(addr)
}

// healthHandler 健康检查处理器
func healthHandler() web.Handler {
	return func(ctx *web.Context) {
		_ = ctx.RespJSON(http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"service": "gateway",
			"message": "Gateway service is healthy",
		})
	}
}

// proxyHandler 代理处理器
func proxyHandler(gatewayService *application.GatewayService) web.Handler {
	return func(ctx *web.Context) {
		// 获取客户端IP
		clientIP := ctx.Req.RemoteAddr
		if clientIP == "" {
			clientIP = "unknown"
		}

		// 读取请求体
		body, err := io.ReadAll(ctx.Req.Body)
		if err != nil {
			logger.Log().Error("httpserver: 读取请求体失败: %v", err)
			_ = ctx.RespJSON(http.StatusBadRequest, map[string]interface{}{"error": "读取请求体失败"})
			return
		}

		// 构建代理请求
		proxyReq := &domain.ProxyRequest{
			Method:  ctx.Req.Method,
			Path:    ctx.Req.URL.Path,
			Headers: ctx.Req.Header,
			Body:    body,
			Client:  clientIP,
		}

		// 如果已鉴权，限流维度改为用户ID（下划线拼接）
		if uid := ctx.Req.Header.Get("X-User-ID"); uid != "" {
			proxyReq.Client = "uid_" + uid
		}

		// 执行代理请求
		reqCtx := ctx.Req.Context()
		resp, err := gatewayService.ProxyRequest(reqCtx, proxyReq)
		if err != nil {
			logger.Log().Error("httpserver: 代理请求失败: %v", err)
			_ = ctx.RespJSON(http.StatusInternalServerError, map[string]interface{}{"error": "代理请求失败"})
			return
		}

		// 设置响应头
		for key, values := range resp.Headers {
			for _, value := range values {
				ctx.Resp.Header().Set(key, value)
			}
		}

		// 设置状态码和响应体
		ctx.RespCode = resp.StatusCode
		ctx.RespData = resp.Body
	}
}
