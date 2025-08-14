package api

import (
	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/errcode"
	"blog-system/common/pkg/logger"
	"blog-system/services/admin/application"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/CoucouMonEcho/go-framework/web"
	webotel "github.com/CoucouMonEcho/go-framework/web/middlewares/opentelemetry"
	webprom "github.com/CoucouMonEcho/go-framework/web/middlewares/prometheus"
)

type HTTPServer struct {
	logger logger.Logger
	server *web.HTTPServer
	app    *application.AdminService
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
	server := web.NewHTTPServer(
		web.ServerWithLogger(lgr.Error),
		web.ServerWithMiddlewares(
			customRecover,
			requestLogger,
			webotel.MiddlewareBuilder{}.Build(),
			webprom.MiddlewareBuilder{Namespace: "blog", Subsystem: "admin", Name: "http", Help: "admin http latency"}.Build(),
		),
	)
	s := &HTTPServer{logger: lgr, server: server}
	s.server.Get("/health", func(ctx *web.Context) {
		_ = ctx.RespJSONOK(dto.Success(map[string]any{"status": "ok", "service": "admin"}))
	})
	// 管理端登录（演示）
	s.server.Post("/api/admin/login", func(ctx *web.Context) {
		var req struct{ Username, Password string }
		if err := ctx.BindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
			_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "参数错误"))
			return
		}
		// 此处应调用 user-service 验证管理员角色，这里简化直接返回
		_ = ctx.RespJSONOK(dto.Success(map[string]any{"token": "mock-admin-token"}))
	})

	// 分级菜单
	// 用户管理
	s.server.Get("/api/admin/users", s.listUsers)
	s.server.Post("/api/admin/users", s.createUser)
	s.server.Post("/api/admin/users/update/:id", s.updateUser)
	s.server.Post("/api/admin/users/delete/:id", s.deleteUser)

	// 文章管理
	s.server.Get("/api/admin/articles", s.listArticles)
	s.server.Post("/api/admin/articles", s.createArticle)
	s.server.Post("/api/admin/articles/update/:id", s.updateArticle)
	s.server.Post("/api/admin/articles/delete/:id", s.deleteArticle)
	// 分类管理
	s.server.Get("/api/admin/categories", s.listCategories)
	s.server.Post("/api/admin/categories", s.createCategory)
	s.server.Post("/api/admin/categories/update/:id", s.updateCategory)
	s.server.Post("/api/admin/categories/delete/:id", s.deleteCategory)
	return s
}

// 简单权限（示例）
func (s *HTTPServer) requireAdmin(ctx *web.Context) bool { return true }

// 具体处理
func (s *HTTPServer) listUsers(ctx *web.Context) {
	_ = ctx.RespJSONOK(dto.Success(map[string]any{"list": []any{}, "total": 0}))
}
func (s *HTTPServer) createUser(ctx *web.Context) { _ = ctx.RespJSONOK(dto.SuccessNil[string]()) }
func (s *HTTPServer) updateUser(ctx *web.Context) { _ = ctx.RespJSONOK(dto.SuccessNil[string]()) }
func (s *HTTPServer) deleteUser(ctx *web.Context) { _ = ctx.RespJSONOK(dto.SuccessNil[string]()) }
func (s *HTTPServer) listArticles(ctx *web.Context) {
	_ = ctx.RespJSONOK(dto.Success(map[string]any{"list": []any{}, "total": 0}))
}
func (s *HTTPServer) createArticle(ctx *web.Context) { _ = ctx.RespJSONOK(dto.SuccessNil[string]()) }
func (s *HTTPServer) updateArticle(ctx *web.Context) { _ = ctx.RespJSONOK(dto.SuccessNil[string]()) }
func (s *HTTPServer) deleteArticle(ctx *web.Context) { _ = ctx.RespJSONOK(dto.SuccessNil[string]()) }
func (s *HTTPServer) listCategories(ctx *web.Context) {
	_ = ctx.RespJSONOK(dto.Success(map[string]any{"list": []any{}, "total": 0}))
}
func (s *HTTPServer) createCategory(ctx *web.Context) { _ = ctx.RespJSONOK(dto.SuccessNil[string]()) }
func (s *HTTPServer) updateCategory(ctx *web.Context) { _ = ctx.RespJSONOK(dto.SuccessNil[string]()) }
func (s *HTTPServer) deleteCategory(ctx *web.Context) { _ = ctx.RespJSONOK(dto.SuccessNil[string]()) }

func (s *HTTPServer) Run(addr string) error { return s.server.Start(addr) }

// SetApp 注入应用服务
func (s *HTTPServer) SetApp(app *application.AdminService) { s.app = app }
