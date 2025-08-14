package api

import (
	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/errcode"
	"blog-system/common/pkg/logger"
	"blog-system/common/pkg/util"
	"blog-system/services/user/application"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/CoucouMonEcho/go-framework/web"
)

// HTTPServer HTTP 服务器
type HTTPServer struct {
	userService *application.UserAppService
	logger      logger.Logger
	server      *web.HTTPServer
}

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(userService *application.UserAppService, logger logger.Logger) *HTTPServer {
	// 恢复中间件
	customRecover := func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()
					logger.LogWithContext(
						"user-service", "recover", "ERROR",
						"panic=%v method=%s path=%s remote=%s\nheaders=%v\nstack=%s",
						r, ctx.Req.Method, ctx.Req.URL.Path, ctx.Req.RemoteAddr, ctx.Req.Header, string(stack),
					)
					_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, "内部服务错误"))
				}
			}()
			next(ctx)
		}
	}
	// 日志中间件
	requestLogger := func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			start := time.Now()
			next(ctx)
			logger.LogWithContext("user-service", "http", "INFO",
				"请求: %s %s %d %s %s",
				ctx.Req.Method, ctx.Req.URL.Path, ctx.RespCode, time.Since(start), ctx.Req.RemoteAddr)
		}
	}

	server := web.NewHTTPServer(
		web.ServerWithLogger(logger.Error),
		web.ServerWithMiddlewares(
			customRecover,
			requestLogger,
		),
	)
	svc := &HTTPServer{
		userService: userService,
		logger:      logger,
		server:      server,
	}
	svc.registerRoutes()
	return svc
}

// registerRoutes 注册路由
func (s *HTTPServer) registerRoutes() {
	// 健康检查
	s.server.Get("/health", s.HealthCheck)

	// 公开接口
	s.server.Post("/api/login", s.Login)

	// 认证接口
	s.server.Use(http.MethodGet, "/api/auth/*", s.AuthMiddleware())
	{
		s.server.Get("/api/auth/info/:user_id", s.GetUserInfo)
		s.server.Post("/api/auth/update", s.UpdateUserInfo)
		s.server.Post("/api/auth/password", s.ChangePassword)
	}
}

// HealthCheck 健康检查
func (s *HTTPServer) HealthCheck(ctx *web.Context) {
	_ = ctx.RespJSONOK(dto.Success(map[string]any{
		"status":    "ok",
		"service":   "user-service",
		"timestamp": time.Now().Unix(),
	}))
}

// Register 接口已下线，注册由 admin 后台负责

// Login 用户登录
func (s *HTTPServer) Login(ctx *web.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	user, err := s.userService.Login(ctx.Req.Context(), req.Username, req.Password)
	if err != nil {
		_ = ctx.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrPasswordInvalid, err.Error()))
		return
	}

	// 生成 JWT 令牌
	token, err := util.GenerateToken(user.ID, user.Role)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}

	_ = ctx.RespJSONOK(dto.Success(map[string]any{
		"token": token,
		"user":  user,
	}))
}

// GetUserInfo 获取用户信息
func (s *HTTPServer) GetUserInfo(ctx *web.Context) {
	userID, err := ctx.PathValue("user_id").AsInt64()
	if err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	user, err := s.userService.GetUserInfo(ctx.Req.Context(), userID)
	if err != nil {
		_ = ctx.RespJSON(http.StatusNotFound, dto.Error(errcode.ErrUserNotFound, err.Error()))
		return
	}

	_ = ctx.RespJSONOK(dto.Success(user))
}

// UpdateUserInfo 更新用户信息（使用 JWT 中的 user_id）
func (s *HTTPServer) UpdateUserInfo(ctx *web.Context) {
	// 从认证中间件写入的上下文中获取用户 ID
	var userID int64
	if v, ok := ctx.UserValues["user_id"]; ok {
		if id, ok2 := v.(int64); ok2 {
			userID = id
		}
	}
	if userID == 0 {
		_ = ctx.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrUnauthorized, "未认证或无效的用户"))
		return
	}

	var req struct {
		Username string `json:"username,omitempty"`
		Email    string `json:"email,omitempty"`
		Avatar   string `json:"avatar,omitempty"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	updates := make(map[string]any)
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}

	if err := s.userService.UpdateUserInfo(ctx.Req.Context(), userID, updates); err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}

	_ = ctx.RespJSONOK(dto.SuccessNil())
}

// ChangePassword 修改密码（使用 JWT 中的 user_id）
func (s *HTTPServer) ChangePassword(ctx *web.Context) {
	var userID int64
	if v, ok := ctx.UserValues["user_id"]; ok {
		if id, ok2 := v.(int64); ok2 {
			userID = id
		}
	}
	if userID == 0 {
		_ = ctx.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrUnauthorized, "未认证或无效的用户"))
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err = ctx.BindJSON(&req); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	if err = s.userService.ChangePassword(ctx.Req.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrPasswordInvalid, err.Error()))
		return
	}

	_ = ctx.RespJSONOK(dto.SuccessNil())
}

// AuthMiddleware JWT 认证中间件
func (s *HTTPServer) AuthMiddleware() web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			token := ctx.Req.Header.Get("Authorization")
			if token == "" {
				_ = ctx.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrUnauthorized, "缺少认证令牌"))
				return
			}

			// 移除 "Bearer " 前缀
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			claims, err := util.ParseToken(token)
			if err != nil {
				_ = ctx.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrTokenInvalid, err.Error()))
				return
			}

			// 存储用户信息
			ctx.UserValues["user_id"] = claims.UserID
			ctx.UserValues["user_role"] = claims.Role

			// next
			next(ctx)
		}
	}
}

// Run 启动服务器
func (s *HTTPServer) Run(addr string) error {
	return s.server.Start(addr)
}
