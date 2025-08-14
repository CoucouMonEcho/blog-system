package api

import (
	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/errcode"
	"blog-system/common/pkg/logger"
	"blog-system/services/stat/infrastructure"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/CoucouMonEcho/go-framework/web"
	webotel "github.com/CoucouMonEcho/go-framework/web/middlewares/opentelemetry"
	webprom "github.com/CoucouMonEcho/go-framework/web/middlewares/prometheus"
)

type HTTPServer struct {
	logger logger.Logger
	server *web.HTTPServer
	repo   *infrastructure.StatRepository
}

func NewHTTPServer(lgr logger.Logger) *HTTPServer {
	customRecover := func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			defer func() {
				if r := recover(); r != nil {
					lgr.LogWithContext("stat-service", "recover", "ERROR", "panic=%v path=%s\nstack=%s", r, ctx.Req.URL.Path, string(debug.Stack()))
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
			lgr.LogWithContext("stat-service", "http", "INFO", "请求: %s %s %d %s", ctx.Req.Method, ctx.Req.URL.Path, ctx.RespCode, time.Since(start))
		}
	}
	server := web.NewHTTPServer(
		web.ServerWithLogger(lgr.Error),
		web.ServerWithMiddlewares(
			customRecover,
			requestLogger,
			webotel.MiddlewareBuilder{}.Build(),
			webprom.MiddlewareBuilder{Namespace: "blog", Subsystem: "stat", Name: "http", Help: "stat http latency"}.Build(),
		),
	)
	s := &HTTPServer{logger: lgr, server: server}
	s.server.Get("/health", func(ctx *web.Context) {
		_ = ctx.RespJSONOK(dto.Success(map[string]any{"status": "ok", "service": "stat"}))
	})
	// 统计 API（简化，不强依赖应用层）
	s.server.Post("/api/stat/incr", s.Incr)
	s.server.Get("/api/stat/get", s.Get)
	return s
}

func (s *HTTPServer) Run(addr string) error { return s.server.Start(addr) }

// SetRepository 注入仓储
func (s *HTTPServer) SetRepository(repo *infrastructure.StatRepository) { s.repo = repo }

// Incr 统计自增: query: type,target_id,target_type[,user_id]
func (s *HTTPServer) Incr(ctx *web.Context) {
	typ := ctx.Req.URL.Query().Get("type")
	targetIDStr := ctx.Req.URL.Query().Get("target_id")
	targetType := ctx.Req.URL.Query().Get("target_type")
	if typ == "" || targetIDStr == "" || targetType == "" {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "缺少必要参数"))
		return
	}
	targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
	if err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "target_id 不合法"))
		return
	}
	var uid *int64
	if s := ctx.Req.URL.Query().Get("user_id"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			uid = &v
		}
	}
	if s.repo == nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, "仓储未初始化"))
		return
	}
	if err := s.repo.Incr(ctx.Req.Context(), typ, targetID, targetType, uid); err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.SuccessNil())
}

// Get 获取统计值
func (s *HTTPServer) Get(ctx *web.Context) {
	typ := ctx.Req.URL.Query().Get("type")
	targetIDStr := ctx.Req.URL.Query().Get("target_id")
	targetType := ctx.Req.URL.Query().Get("target_type")
	if typ == "" || targetIDStr == "" || targetType == "" {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "缺少必要参数"))
		return
	}
	targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
	if err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "target_id 不合法"))
		return
	}
	var uid *int64
	if s := ctx.Req.URL.Query().Get("user_id"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			uid = &v
		}
	}
	if s.repo == nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, "仓储未初始化"))
		return
	}
	val, err := s.repo.Get(ctx.Req.Context(), typ, targetID, targetType, uid)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(map[string]any{"value": val}))
}
