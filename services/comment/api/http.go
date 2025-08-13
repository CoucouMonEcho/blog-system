package api

import (
	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/errcode"
	"blog-system/common/pkg/logger"
	"blog-system/services/comment/application"
	"blog-system/services/comment/domain"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/CoucouMonEcho/go-framework/web"
)

type HTTPServer struct {
	svc    *application.CommentAppService
	logger logger.Logger
	server *web.HTTPServer
}

func NewHTTPServer(svc *application.CommentAppService, logger logger.Logger) *HTTPServer {
	customRecover := func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			defer func() {
				if r := recover(); r != nil {
					logger.LogWithContext("comment-service", "recover", "ERROR", "panic=%v method=%s path=%s remote=%s\nheaders=%v\nstack=%s", r, ctx.Req.Method, ctx.Req.URL.Path, ctx.Req.RemoteAddr, ctx.Req.Header, string(debug.Stack()))
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
			logger.LogWithContext("comment-service", "http", "INFO", "请求: %s %s %d %s %s", ctx.Req.Method, ctx.Req.URL.Path, ctx.RespCode, time.Since(start), ctx.Req.RemoteAddr)
		}
	}
	server := web.NewHTTPServer(web.ServerWithLogger(logger.Error), web.ServerWithMiddlewares(customRecover, requestLogger))
	s := &HTTPServer{svc: svc, logger: logger, server: server}
	s.registerRoutes()
	return s
}

func (s *HTTPServer) registerRoutes() {
	s.server.Get("/health", func(ctx *web.Context) {
		_ = ctx.RespJSONOK(dto.Success(map[string]any{"status": "ok", "service": "comment"}))
	})
	s.server.Post("/api/comment", s.Create)
	s.server.Get("/api/comment/:id", s.Get)
}

func (s *HTTPServer) Create(ctx *web.Context) {
	var req struct {
		Content   string `json:"content" binding:"required"`
		UserID    int64  `json:"user_id" binding:"required"`
		ArticleID int64  `json:"article_id" binding:"required"`
		ParentID  int64  `json:"parent_id"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}
	c := &domain.Comment{Content: req.Content, UserID: req.UserID, ArticleID: req.ArticleID, ParentID: req.ParentID, Status: 1}
	res, err := s.svc.Create(ctx.Req.Context(), c)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(res))
}

func (s *HTTPServer) Get(ctx *web.Context) {
	id, err := ctx.PathValue("id").AsInt64()
	if err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}
	res, err := s.svc.GetByID(ctx.Req.Context(), id)
	if err != nil {
		_ = ctx.RespJSON(http.StatusNotFound, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(res))
}

func (s *HTTPServer) Run(addr string) error { return s.server.Start(addr) }
