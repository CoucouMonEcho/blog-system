package api

import (
	"io"
	"net/http"
	"time"

	"blog-system/common/pkg/logger"
	"blog-system/services/gateway/application"
	"blog-system/services/gateway/domain"

	"github.com/CoucouMonEcho/go-framework/web"
	"github.com/CoucouMonEcho/go-framework/web/middlewares/recover"
)

// HTTPServer HTTP服务器
type HTTPServer struct {
	gatewayService *application.GatewayService
	server         *web.HTTPServer
}

// NewHTTPServer 创建HTTP服务器
func NewHTTPServer(gatewayService *application.GatewayService) *HTTPServer {
	// Request ID 中间件
	requestIDMiddleware := func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			requestID := time.Now().Format("20060102150405.000000")
			ctx.Resp.Header().Set("X-Request-ID", requestID)
			next(ctx)
		}
	}

	// 使用go-framework的recover中间件
	recoverMiddleware := recover.MiddlewareBuilder{
		Code: http.StatusInternalServerError,
		Data: []byte("Internal Server Error"),
		Log: func(ctx *web.Context) {
			logger.Log().Error("recover: 服务恢复中间件触发")
		},
	}.Build()

	server := web.NewHTTPServer(
		web.ServerWithLogger(logger.Log().Error),
		web.ServerWithMiddlewares(
			requestIDMiddleware,
			recoverMiddleware,
			loggerMiddleware(),
			corsMiddleware(),
		),
	)

	return &HTTPServer{
		gatewayService: gatewayService,
		server:         server,
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

// loggerMiddleware 日志中间件
func loggerMiddleware() web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			start := time.Now()

			next(ctx)

			logger.Log().Info("http: 请求: %s %s %d %s %s", ctx.Req.Method, ctx.Req.URL.Path, ctx.RespCode, time.Since(start), ctx.Req.RemoteAddr)
		}
	}
}

// corsMiddleware CORS中间件
func corsMiddleware() web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			ctx.Resp.Header().Set("Access-Control-Allow-Origin", "*")
			ctx.Resp.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			ctx.Resp.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			if ctx.Req.Method == "OPTIONS" {
				_ = ctx.RespJSON(http.StatusNoContent, "")
				return
			}

			next(ctx)
		}
	}
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
			logger.Log().Error("proxy: 读取请求体失败: %v", err)
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

		// 执行代理请求
		reqCtx := ctx.Req.Context()
		resp, err := gatewayService.ProxyRequest(reqCtx, proxyReq)
		if err != nil {
			logger.Log().Error("proxy: 代理请求失败: %v", err)
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
