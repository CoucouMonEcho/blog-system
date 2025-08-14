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
	server *web.HTTPServer
	repo   *infrastructure.StatRepository
	agg    *infrastructure.PVAggregator
}

func NewHTTPServer() *HTTPServer {
	// Request ID 中间件
	requestIDMiddleware := func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			requestID := logger.GenerateRequestID()
			ctx.Req = ctx.Req.WithContext(logger.WithRequestID(ctx.Req.Context(), requestID))
			ctx.Resp.Header().Set("X-Request-ID", requestID)
			next(ctx)
		}
	}

	customRecover := func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			defer func() {
				if r := recover(); r != nil {
					logger.L().LogWithContextAndRequestID(ctx.Req.Context(), "stat-service", "recover", "ERROR", "panic=%v path=%s\nstack=%s", r, ctx.Req.URL.Path, string(debug.Stack()))
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
			logger.L().LogWithContextAndRequestID(ctx.Req.Context(), "stat-service", "http", "INFO", "请求: %s %s %d %s", ctx.Req.Method, ctx.Req.URL.Path, ctx.RespCode, time.Since(start))
		}
	}
	server := web.NewHTTPServer(
		web.ServerWithLogger(logger.L().Error),
		web.ServerWithMiddlewares(
			requestIDMiddleware,
			customRecover,
			requestLogger,
			webotel.MiddlewareBuilder{}.Build(),
			webprom.MiddlewareBuilder{Namespace: "blog", Subsystem: "stat", Name: "http", Help: "stat http latency"}.Build(),
		),
	)
	s := &HTTPServer{server: server, agg: infrastructure.NewPVAggregator()}
	s.server.Get("/health", func(ctx *web.Context) {
		_ = ctx.RespJSONOK(dto.Success(map[string]any{"status": "ok", "service": "stat"}))
	})
	// 统计 API（简化，不强依赖应用层）
	s.server.Post("/api/incr", s.Incr)
	s.server.Get("/api/get", s.Get)
	// 仪表盘占位 API
	s.server.Get("/api/stat/overview", s.Overview)
	s.server.Get("/api/stat/pv_timeseries", s.PVTimeSeries)
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
	// 记录到内存聚合器（PV/UV估算）
	s.agg.RecordPV(time.Now(), uid)
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

// Overview 仪表盘总览占位：pv_today, uv_today, online_users, article_total, category_total, error_5xx_last_1h
func (s *HTTPServer) Overview(ctx *web.Context) {
	pv, uv, online := s.agg.Overview(time.Now())
	// 其他值占位为 0，由 admin 聚合调用 content 统计实际值
	_ = ctx.RespJSONOK(dto.Success(map[string]any{
		"pv_today":          pv,
		"uv_today":          uv,
		"online_users":      online,
		"article_total":     0,
		"category_total":    0,
		"error_5xx_last_1h": 0,
	}))
}

// PVTimeSeries 返回时间序列，interval 支持 5m/1h/1d
func (s *HTTPServer) PVTimeSeries(ctx *web.Context) {
	from, to, step, ok := parseRange(ctx)
	if !ok {
		return
	}
	series := s.agg.PVTimeSeries(from, to, step)
	_ = ctx.RespJSONOK(dto.Success(series))
}

// parseRange 统一 from,to,interval 解析
func parseRange(ctx *web.Context) (time.Time, time.Time, time.Duration, bool) {
	q := ctx.Req.URL.Query()
	fromStr := q.Get("from")
	toStr := q.Get("to")
	intervalStr := q.Get("interval")
	if fromStr == "" || toStr == "" || intervalStr == "" {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "缺少必要参数"))
		return time.Time{}, time.Time{}, 0, false
	}
	from, err1 := time.Parse(time.RFC3339, fromStr)
	to, err2 := time.Parse(time.RFC3339, toStr)
	var step time.Duration
	switch intervalStr {
	case "5m":
		step = 5 * time.Minute
	case "1h":
		step = time.Hour
	case "1d":
		step = 24 * time.Hour
	default:
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "interval 仅支持 5m/1h/1d"))
		return time.Time{}, time.Time{}, 0, false
	}
	if err1 != nil || err2 != nil || from.After(to) {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "时间范围不合法"))
		return time.Time{}, time.Time{}, 0, false
	}
	return from, to, step, true
}
