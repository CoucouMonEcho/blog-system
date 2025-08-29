package httpserver

import (
	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/errcode"
	"blog-system/common/pkg/logger"
	"blog-system/services/content/application"
	"blog-system/services/content/domain"
	"net/http"
	"strconv"
	"time"

	"github.com/CoucouMonEcho/go-framework/web"
	"github.com/CoucouMonEcho/go-framework/web/middlewares/accesslog"
	"github.com/CoucouMonEcho/go-framework/web/middlewares/errhandle"
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
			ctx.Resp.Header().Set("X-Request-ID", time.Now().Format("20060102150405.000000"))
			next(ctx)
		}
	}

	server := web.NewHTTPServer(
		web.ServerWithLogger(logger.Log().Error),
		web.ServerWithMiddlewares(
			requestIDMiddleware,
			errhandle.NewMiddlewareBuilder().RegisterError(http.StatusInternalServerError, []byte("内部服务错误")).Build(),
			accesslog.NewMiddlewareBuilder().LogFunc(func(log string) { logger.Log().Info(log) }).Build(),
			webprom.MiddlewareBuilder{Namespace: "blog-system", Subsystem: "content", Name: "http", Help: "content http latency"}.Build(),
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
	s.server.Get("/api/category/list", s.ListCategories)
}

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

// ListCategories 返回分类全量列表
func (s *HTTPServer) ListCategories(ctx *web.Context) {
	list, err := s.contentService.ListAllCategories(ctx.Req.Context())
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(list))
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
