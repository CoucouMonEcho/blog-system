package api

import (
	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/logger"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/CoucouMonEcho/go-framework/web"
)

type HTTPServer struct {
	logger logger.Logger
	server *web.HTTPServer
}

func NewHTTPServer(lgr logger.Logger) *HTTPServer {
	customRecover := func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			defer func() {
				if r := recover(); r != nil {
					lgr.LogWithContext("admin-service", "recover", "ERROR", "panic=%v path=%s\nstack=%s", r, ctx.Req.URL.Path, string(debug.Stack()))
					_ = ctx.RespJSON(http.StatusInternalServerError, map[string]any{"error": "内部服务错误"})
				}
			}()
			next(ctx)
		}
	}
	requestLogger := func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			start := time.Now()
			next(ctx)
			lgr.LogWithContext("admin-service", "http", "INFO", "请求: %s %s %d %s", ctx.Req.Method, ctx.Req.URL.Path, ctx.RespCode, time.Since(start))
		}
	}
	server := web.NewHTTPServer(web.ServerWithLogger(lgr.Error), web.ServerWithMiddlewares(customRecover, requestLogger))
	s := &HTTPServer{logger: lgr, server: server}
	s.server.Get("/health", func(ctx *web.Context) {
		_ = ctx.RespJSONOK(dto.Success(map[string]any{"status": "ok", "service": "admin"}))
	})
	return s
}

func (s *HTTPServer) Run(addr string) error { return s.server.Start(addr) }
