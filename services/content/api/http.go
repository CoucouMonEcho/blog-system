package api

import (
	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/errcode"
	"blog-system/common/pkg/logger"
	"blog-system/services/content/application"
	"blog-system/services/content/domain"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/CoucouMonEcho/go-framework/web"
)

type HTTPServer struct {
	contentService *application.ContentAppService
	logger         logger.Logger
	server         *web.HTTPServer
}

func NewHTTPServer(contentService *application.ContentAppService, logger logger.Logger) *HTTPServer {
	customRecover := func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			defer func() {
				if r := recover(); r != nil {
					logger.LogWithContext("content-service", "recover", "ERROR", "panic=%v method=%s path=%s remote=%s\nheaders=%v\nstack=%s", r, ctx.Req.Method, ctx.Req.URL.Path, ctx.Req.RemoteAddr, ctx.Req.Header, string(debug.Stack()))
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
			logger.LogWithContext("content-service", "http", "INFO", "请求: %s %s %d %s %s", ctx.Req.Method, ctx.Req.URL.Path, ctx.RespCode, time.Since(start), ctx.Req.RemoteAddr)
		}
	}

	server := web.NewHTTPServer(
		web.ServerWithLogger(logger.Error),
		web.ServerWithMiddlewares(customRecover, requestLogger),
	)

	s := &HTTPServer{contentService: contentService, logger: logger, server: server}
	s.registerRoutes()
	return s
}

func (s *HTTPServer) registerRoutes() {
	s.server.Get("/health", func(ctx *web.Context) {
		_ = ctx.RespJSONOK(dto.Success(map[string]any{"status": "ok", "service": "content"}))
	})

	// 基础 API
	s.server.Post("/api/article", s.CreateArticle)
	s.server.Get("/api/article/:id", s.GetArticle)
}

func (s *HTTPServer) CreateArticle(ctx *web.Context) {
	var req struct {
		Title      string `json:"title" binding:"required"`
		Slug       string `json:"slug" binding:"required"`
		Content    string `json:"content" binding:"required"`
		Summary    string `json:"summary"`
		AuthorID   int64  `json:"author_id" binding:"required"`
		CategoryID int64  `json:"category_id" binding:"required"`
		Status     int    `json:"status"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}
	art := &domain.Article{
		Title: req.Title, Slug: req.Slug, Content: req.Content, Summary: req.Summary,
		AuthorID: req.AuthorID, CategoryID: req.CategoryID, Status: req.Status,
	}
	res, err := s.contentService.Create(ctx.Req.Context(), art)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(res))
}

func (s *HTTPServer) GetArticle(ctx *web.Context) {
	id, err := ctx.PathValue("id").AsInt64()
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

func (s *HTTPServer) Run(addr string) error { return s.server.Start(addr) }
