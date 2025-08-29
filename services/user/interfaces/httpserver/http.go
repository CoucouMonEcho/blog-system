package httpserver

import (
	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/errcode"
	"blog-system/common/pkg/logger"
	"blog-system/services/user/application"
	"net/http"
	"time"

	"github.com/CoucouMonEcho/go-framework/web"
	"github.com/CoucouMonEcho/go-framework/web/middlewares/accesslog"
	"github.com/CoucouMonEcho/go-framework/web/middlewares/errhandle"
	webprom "github.com/CoucouMonEcho/go-framework/web/middlewares/prometheus"
)

// HTTPServer HTTP 服务器
type HTTPServer struct {
	userService *application.UserAppService
	server      *web.HTTPServer
}

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(userService *application.UserAppService) *HTTPServer {
	svc := &HTTPServer{
		userService: userService,
		server: web.NewHTTPServer(
			web.ServerWithLogger(logger.Log().Error),
			web.ServerWithMiddlewares(
				errhandle.NewMiddlewareBuilder().RegisterError(http.StatusInternalServerError, []byte("内部服务错误")).Build(),
				accesslog.NewMiddlewareBuilder().LogFunc(func(log string) { logger.Log().Info(log) }).Build(),
				webprom.MiddlewareBuilder{Namespace: "blog-system", Subsystem: "user", Name: "http", Help: "user http latency"}.Build(),
			),
		),
	}
	svc.registerRoutes()
	return svc
}

// registerRoutes 注册路由
func (s *HTTPServer) registerRoutes() {
	s.server.Get("/health", s.HealthCheck)
	s.server.Post("/api/login", s.Login)
	s.server.Get("/api/info/:user_id", s.GetUserInfo)
	s.server.Post("/api/update", s.UpdateUserInfo)
	s.server.Post("/api/password", s.ChangePassword)
}

// HealthCheck 健康检查
func (s *HTTPServer) HealthCheck(ctx *web.Context) {
	_ = ctx.RespJSONOK(dto.Success(map[string]any{
		"status":    "ok",
		"service":   "user-service",
		"timestamp": time.Now().Unix(),
	}))
}

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
	user, token, err := s.userService.Login(ctx.Req.Context(), req.Username, req.Password)
	if err != nil {
		_ = ctx.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrPasswordInvalid, err.Error()))
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
	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}
	if err := s.userService.ChangePassword(ctx.Req.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrPasswordInvalid, err.Error()))
		return
	}
	_ = ctx.RespJSONOK(dto.SuccessNil())
}

// Run 启动服务器
func (s *HTTPServer) Run(addr string) error {
	return s.server.Start(addr)
}
