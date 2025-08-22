package httpserver

import (
	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/errcode"
	"blog-system/common/pkg/logger"
	"blog-system/services/content/application"
	"blog-system/services/content/domain"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/CoucouMonEcho/go-framework/web"
	webotel "github.com/CoucouMonEcho/go-framework/web/middlewares/opentelemetry"
	webprom "github.com/CoucouMonEcho/go-framework/web/middlewares/prometheus"
)

type HTTPServer struct {
	contentService *application.ContentAppService
	server         *web.HTTPServer
}

func NewHTTPServer(contentService *application.ContentAppService) *HTTPServer {
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
					logger.L().LogWithContextAndRequestID(ctx.Req.Context(), "content-service", "recover", "ERROR", "panic=%v method=%s path=%s remote=%s\nheaders=%v\nstack=%s", r, ctx.Req.Method, ctx.Req.URL.Path, ctx.Req.RemoteAddr, ctx.Req.Header, string(debug.Stack()))
					_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, "内部服务错误"))
				}
			}()
			next(ctx)
		}
	}

	requestLogger := func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			start := time.Now()
			next(ctx)
			logger.L().LogWithContextAndRequestID(ctx.Req.Context(), "content-service", "http", "INFO", "请求: %s %s %d %s %s", ctx.Req.Method, ctx.Req.URL.Path, ctx.RespCode, time.Since(start), ctx.Req.RemoteAddr)
		}
	}

	server := web.NewHTTPServer(
		web.ServerWithLogger(logger.L().Error),
		web.ServerWithMiddlewares(
			requestIDMiddleware,
			customRecover,
			requestLogger,
			webotel.MiddlewareBuilder{}.Build(),
			webprom.MiddlewareBuilder{Namespace: "blog", Subsystem: "content", Name: "http", Help: "content http latency"}.Build(),
		),
	)

	s := &HTTPServer{contentService: contentService, server: server}
	s.registerRoutes()
	return s
}

func (s *HTTPServer) registerRoutes() {
	s.server.Get("/health", func(ctx *web.Context) {
		_ = ctx.RespJSONOK(dto.Success(map[string]any{"status": "ok", "service": "content"}))
	})

	// 只保留只读接口：列表与详情
	s.server.Get("/api/article/:article_id", s.GetArticle)
	s.server.Get("/api/article/list", s.ListArticleSummaries)
	s.server.Get("/api/article/search", s.SearchArticles)
	s.server.Get("/api/category/tree", s.ListCategoryTree)
}

// Create/Update/Delete 移至 admin 模块

func (s *HTTPServer) GetArticle(ctx *web.Context) {
	id, err := ctx.PathValue("article_id").AsInt64()
	if err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}
	art, err := s.contentService.GetByID(ctx.Req.Context(), id)
	if err != nil {
		_ = ctx.RespJSON(http.StatusNotFound, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(art))
}

// UpdateArticle 更新文章
// UpdateArticle 移至 admin 模块

// DeleteArticle 删除文章
// DeleteArticle 移至 admin 模块

// ListArticleSummaries 文章列表（ID+Title）
func (s *HTTPServer) ListArticleSummaries(ctx *web.Context) {
	page, pageSize := parsePagination(ctx)
	list, total, err := s.contentService.ListSummaries(ctx.Req.Context(), page, pageSize)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(dto.PageResponse[*domain.ArticleSummary]{
		List: list, Total: total, Page: page, PageSize: pageSize,
	}))
}

// SearchArticles 根据内容全文搜索文章（返回摘要列表）
func (s *HTTPServer) SearchArticles(ctx *web.Context) {
	q := ctx.Req.URL.Query().Get("q")
	if q == "" {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "缺少查询关键词 q"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := s.contentService.SearchSummaries(ctx.Req.Context(), q, page, pageSize)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(dto.PageResponse[*domain.ArticleSummary]{List: list, Total: total, Page: page, PageSize: pageSize}))
}

// ListCategoryTree 返回树状三级分类
func (s *HTTPServer) ListCategoryTree(ctx *web.Context) {
	tree, err := s.contentService.GetCategoryTree(ctx.Req.Context())
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(tree))
}

// parsePagination 统一分页解析
func parsePagination(ctx *web.Context) (int, int) {
	page := 1
	pageSize := 10
	if v := ctx.Req.URL.Query().Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v := ctx.Req.URL.Query().Get("page_size"); v != "" {
		if ps, err := strconv.Atoi(v); err == nil && ps > 0 {
			pageSize = ps
		}
	}
	return page, pageSize
}

func (s *HTTPServer) Run(addr string) error { return s.server.Start(addr) }
