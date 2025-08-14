package api

import (
	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/errcode"
	"blog-system/common/pkg/logger"
	"blog-system/common/pkg/util"
	"blog-system/services/admin/application"
	"blog-system/services/admin/domain"
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
	app    *application.AdminService
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
					logger.L().LogWithContextAndRequestID(ctx.Req.Context(), "admin-service", "recover", "ERROR", "panic=%v path=%s\nstack=%s", r, ctx.Req.URL.Path, string(debug.Stack()))
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
			logger.L().LogWithContextAndRequestID(ctx.Req.Context(), "admin-service", "http", "INFO", "请求: %s %s %d %s", ctx.Req.Method, ctx.Req.URL.Path, ctx.RespCode, time.Since(start))
		}
	}
	server := web.NewHTTPServer(
		web.ServerWithLogger(logger.L().Error),
		web.ServerWithMiddlewares(
			requestIDMiddleware,
			customRecover,
			requestLogger,
			webotel.MiddlewareBuilder{}.Build(),
			webprom.MiddlewareBuilder{Namespace: "blog", Subsystem: "admin", Name: "http", Help: "admin http latency"}.Build(),
		),
	)
	s := &HTTPServer{server: server}
	s.server.Get("/health", func(ctx *web.Context) {
		_ = ctx.RespJSONOK(dto.Success(map[string]any{"status": "ok", "service": "admin"}))
	})
	// 管理端登录（调用 user-service）
	s.server.Post("/api/login", func(ctx *web.Context) {
		var req struct{ Username, Password string }
		if err := ctx.BindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
			_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "参数错误"))
			return
		}
		if s.app == nil {
			_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, "应用未初始化"))
			return
		}
		token, err := s.app.AdminLogin(ctx.Req.Context(), req.Username, req.Password)
		if err != nil {
			_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
			return
		}
		_ = ctx.RespJSONOK(dto.Success(map[string]any{"token": token}))
	})
	// 认证拦截（除登录外的所有 /api/* 路由）
	s.server.Use(http.MethodGet, "/api/*", s.adminAuth())
	s.server.Use(http.MethodPost, "/api/*", s.adminAuth())

	// 用户管理
	s.server.Get("/api/users", s.listUsers)
	s.server.Post("/api/users", s.createUser)
	s.server.Post("/api/users/update/:id", s.updateUser)
	s.server.Post("/api/users/delete/:id", s.deleteUser)

	// 文章管理
	s.server.Get("/api/articles", s.listArticles)
	s.server.Post("/api/articles", s.createArticle)
	s.server.Post("/api/articles/update/:id", s.updateArticle)
	s.server.Post("/api/articles/delete/:id", s.deleteArticle)

	// 分类管理
	s.server.Get("/api/categories", s.listCategories)
	s.server.Post("/api/categories", s.createCategory)
	s.server.Post("/api/categories/update/:id", s.updateCategory)
	s.server.Post("/api/categories/delete/:id", s.deleteCategory)
	// 分级菜单（树）
	s.server.Get("/api/categories/tree", s.categoryTree)
	// 仪表盘统计
	s.server.Get("/api/stat/overview", s.statOverview)
	s.server.Get("/api/stat/pv_timeseries", s.statPVSeries)
	s.server.Get("/api/stat/error_rate", s.statErrorRate)
	s.server.Get("/api/stat/latency_percentile", s.statLatency)
	s.server.Get("/api/stat/top_endpoints", s.statTopEndpoints)
	s.server.Get("/api/stat/active_users", s.statActiveUsers)
	return s
}

// 简单权限检查：基于 Authorization JWT 的 role == "admin"
func (s *HTTPServer) requireAdmin(ctx *web.Context) bool {
	token := ctx.Req.Header.Get("Authorization")
	if token == "" {
		_ = ctx.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrUnauthorized, "缺少认证令牌"))
		return false
	}
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	claims, err := util.ParseToken(token)
	if err != nil {
		_ = ctx.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrTokenInvalid, err.Error()))
		return false
	}
	if claims.Role != "admin" {
		_ = ctx.RespJSON(http.StatusForbidden, dto.Error(errcode.ErrAdminForbidden, "需要管理员权限"))
		return false
	}
	ctx.UserValues["admin_user_id"] = claims.UserID
	return true
}

// adminAuth 作为中间件应用到 /api/*（跳过 /api/login）
func (s *HTTPServer) adminAuth() web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			if ctx.Req.URL.Path == "/api/login" {
				next(ctx)
				return
			}
			if !s.requireAdmin(ctx) {
				return
			}
			next(ctx)
		}
	}
}

// 具体处理
func (s *HTTPServer) listUsers(ctx *web.Context) {
	page, pageSize := parsePagination(ctx)
	list, total, err := s.app.ListUsers(ctx.Req.Context(), page, pageSize)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(dto.PageResponse[*domain.User]{List: list, Total: total, Page: page, PageSize: pageSize}))
}

func (s *HTTPServer) createUser(ctx *web.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Avatar   string `json:"avatar"`
		Status   int    `json:"status"`
	}
	if err := ctx.BindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "参数错误"))
		return
	}
	u := &domain.User{Username: req.Username, Email: req.Email, Password: req.Password, Role: req.Role, Avatar: req.Avatar, Status: req.Status}
	if err := s.app.CreateUser(ctx.Req.Context(), u); err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.SuccessNil())
}

func (s *HTTPServer) updateUser(ctx *web.Context) {
	id, err := ctx.PathValue("id").AsInt64()
	if err != nil || id <= 0 {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "id 不合法"))
		return
	}
	var req struct {
		Username string `json:"username,omitempty"`
		Email    string `json:"email,omitempty"`
		Password string `json:"password,omitempty"`
		Role     string `json:"role,omitempty"`
		Avatar   string `json:"avatar,omitempty"`
		Status   int    `json:"status,omitempty"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}
	u := &domain.User{ID: id, Username: req.Username, Email: req.Email, Password: req.Password, Role: req.Role, Avatar: req.Avatar, Status: req.Status}
	if err := s.app.UpdateUser(ctx.Req.Context(), u); err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.SuccessNil())
}

func (s *HTTPServer) deleteUser(ctx *web.Context) {
	id, err := ctx.PathValue("id").AsInt64()
	if err != nil || id <= 0 {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "id 不合法"))
		return
	}
	if err := s.app.DeleteUser(ctx.Req.Context(), id); err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.SuccessNil())
}

func (s *HTTPServer) listArticles(ctx *web.Context) {
	page, pageSize := parsePagination(ctx)
	list, total, err := s.app.ListArticles(ctx.Req.Context(), page, pageSize)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(dto.PageResponse[*domain.Article]{List: list, Total: total, Page: page, PageSize: pageSize}))
}

func (s *HTTPServer) createArticle(ctx *web.Context) {
	var req struct {
		Title       string  `json:"title"`
		Slug        string  `json:"slug"`
		Content     string  `json:"content"`
		Summary     string  `json:"summary"`
		AuthorID    int64   `json:"author_id"`
		CategoryID  int64   `json:"category_id"`
		Status      int     `json:"status"`
		IsTop       bool    `json:"is_top"`
		IsRecommend bool    `json:"is_recommend"`
		PublishedAt *string `json:"published_at,omitempty"`
	}
	if err := ctx.BindJSON(&req); err != nil || req.Title == "" {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "参数错误"))
		return
	}
	a := &domain.Article{
		Title: req.Title, Slug: req.Slug, Content: req.Content, Summary: req.Summary,
		AuthorID: req.AuthorID, CategoryID: req.CategoryID, Status: req.Status,
		IsTop: req.IsTop, IsRecommend: req.IsRecommend,
	}
	if err := s.app.CreateArticle(ctx.Req.Context(), a); err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.SuccessNil())
}

func (s *HTTPServer) updateArticle(ctx *web.Context) {
	id, err := ctx.PathValue("id").AsInt64()
	if err != nil || id <= 0 {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "id 不合法"))
		return
	}
	var req struct {
		Title       string `json:"title,omitempty"`
		Slug        string `json:"slug,omitempty"`
		Content     string `json:"content,omitempty"`
		Summary     string `json:"summary,omitempty"`
		CategoryID  int64  `json:"category_id,omitempty"`
		Status      int    `json:"status,omitempty"`
		IsTop       bool   `json:"is_top,omitempty"`
		IsRecommend bool   `json:"is_recommend,omitempty"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}
	a := &domain.Article{ID: id, Title: req.Title, Slug: req.Slug, Content: req.Content, Summary: req.Summary, CategoryID: req.CategoryID, Status: req.Status, IsTop: req.IsTop, IsRecommend: req.IsRecommend}
	if err := s.app.UpdateArticle(ctx.Req.Context(), a); err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.SuccessNil())
}

func (s *HTTPServer) deleteArticle(ctx *web.Context) {
	id, err := ctx.PathValue("id").AsInt64()
	if err != nil || id <= 0 {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "id 不合法"))
		return
	}
	if err := s.app.DeleteArticle(ctx.Req.Context(), id); err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.SuccessNil())
}

func (s *HTTPServer) listCategories(ctx *web.Context) {
	page, pageSize := parsePagination(ctx)
	list, total, err := s.app.ListCategories(ctx.Req.Context(), page, pageSize)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(dto.PageResponse[*domain.Category]{List: list, Total: total, Page: page, PageSize: pageSize}))
}

func (s *HTTPServer) createCategory(ctx *web.Context) {
	var req struct {
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		ParentID int64  `json:"parent_id"`
		Sort     int    `json:"sort"`
	}
	if err := ctx.BindJSON(&req); err != nil || req.Name == "" {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "参数错误"))
		return
	}
	c := &domain.Category{Name: req.Name, Slug: req.Slug, ParentID: req.ParentID, Sort: req.Sort}
	if err := s.app.CreateCategory(ctx.Req.Context(), c); err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.SuccessNil())
}

func (s *HTTPServer) updateCategory(ctx *web.Context) {
	id, err := ctx.PathValue("id").AsInt64()
	if err != nil || id <= 0 {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "id 不合法"))
		return
	}
	var req struct {
		Name     string `json:"name,omitempty"`
		Slug     string `json:"slug,omitempty"`
		ParentID int64  `json:"parent_id,omitempty"`
		Sort     int    `json:"sort,omitempty"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}
	c := &domain.Category{ID: id, Name: req.Name, Slug: req.Slug, ParentID: req.ParentID, Sort: req.Sort}
	if err := s.app.UpdateCategory(ctx.Req.Context(), c); err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.SuccessNil())
}

func (s *HTTPServer) deleteCategory(ctx *web.Context) {
	id, err := ctx.PathValue("id").AsInt64()
	if err != nil || id <= 0 {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "id 不合法"))
		return
	}
	if err := s.app.DeleteCategory(ctx.Req.Context(), id); err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.SuccessNil())
}

// categoryTree 构建树状分类（最多三级，按 Sort 升序）
func (s *HTTPServer) categoryTree(ctx *web.Context) {
	// 拉取足够多的数据用于构建树
	list, _, err := s.app.ListCategories(ctx.Req.Context(), 1, 10000)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	type Node struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		ParentID int64  `json:"parent_id"`
		Sort     int    `json:"sort"`
		Children []Node `json:"children,omitempty"`
	}
	// 索引
	idToNode := make(map[int64]*Node)
	for _, c := range list {
		idToNode[c.ID] = &Node{ID: c.ID, Name: c.Name, Slug: c.Slug, ParentID: c.ParentID, Sort: c.Sort}
	}
	// 依据原始顺序（已按 Sort 排序）组装树
	roots := make([]*Node, 0)
	for _, c := range list {
		n := idToNode[c.ID]
		if c.ParentID == 0 {
			roots = append(roots, n)
			continue
		}
		if p, ok := idToNode[c.ParentID]; ok {
			p.Children = append(p.Children, *n)
		} else {
			roots = append(roots, n)
		}
	}
	// 转换为值切片
	out := make([]Node, 0, len(roots))
	for _, r := range roots {
		out = append(out, *r)
	}
	_ = ctx.RespJSONOK(dto.Success(out))
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

// 仪表盘
func (s *HTTPServer) statOverview(ctx *web.Context) {
	data, err := s.app.Dashboard(ctx.Req.Context())
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(data))
}

func (s *HTTPServer) statPVSeries(ctx *web.Context) {
	q := ctx.Req.URL.Query()
	from := q.Get("from")
	to := q.Get("to")
	interval := q.Get("interval")
	if from == "" || to == "" || interval == "" {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "缺少必要参数"))
		return
	}
	series, err := s.app.PVSeries(ctx.Req.Context(), from, to, interval)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(series))
}

func (s *HTTPServer) statErrorRate(ctx *web.Context) {
	q := ctx.Req.URL.Query()
	from, to, service := q.Get("from"), q.Get("to"), q.Get("service")
	if from == "" || to == "" || service == "" {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "缺少必要参数"))
		return
	}
	val, err := s.app.ErrorRate(ctx.Req.Context(), from, to, service)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(map[string]any{"error_rate": val}))
}

func (s *HTTPServer) statLatency(ctx *web.Context) {
	q := ctx.Req.URL.Query()
	from, to, service := q.Get("from"), q.Get("to"), q.Get("service")
	if from == "" || to == "" || service == "" {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "缺少必要参数"))
		return
	}
	val, err := s.app.LatencyPercentile(ctx.Req.Context(), from, to, service)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(val))
}

func (s *HTTPServer) statTopEndpoints(ctx *web.Context) {
	q := ctx.Req.URL.Query()
	from, to, service := q.Get("from"), q.Get("to"), q.Get("service")
	if from == "" || to == "" || service == "" {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "缺少必要参数"))
		return
	}
	topN := 10
	if v := q.Get("top"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			topN = n
		}
	}
	vals, err := s.app.TopEndpoints(ctx.Req.Context(), from, to, service, topN)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(vals))
}

func (s *HTTPServer) statActiveUsers(ctx *web.Context) {
	q := ctx.Req.URL.Query()
	from, to := q.Get("from"), q.Get("to")
	if from == "" || to == "" {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, "缺少必要参数"))
		return
	}
	val, err := s.app.ActiveUsers(ctx.Req.Context(), from, to)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.Success(map[string]any{"active_users": val}))
}

func (s *HTTPServer) Run(addr string) error { return s.server.Start(addr) }

// SetApp 注入应用服务
func (s *HTTPServer) SetApp(app *application.AdminService) { s.app = app }
