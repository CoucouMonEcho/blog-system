package api

import (
	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/errcode"
	"blog-system/common/pkg/util"
	"blog-system/services/user/application"
	"net/http"
	"time"

	"github.com/CoucouMonEcho/go-framework/web"
)

// Controller HTTP 服务器
type Controller struct {
	userService *application.UserAppService
	server      *web.HTTPServer
}

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(userService *application.UserAppService) *Controller {
	//TODO middleware
	server := web.NewHTTPServer()
	controller := &Controller{
		userService: userService,
		server:      server,
	}
	controller.registerRoutes()
	return controller
}

// registerRoutes 注册路由
func (c *Controller) registerRoutes() {
	// 健康检查
	c.server.Get("/health", c.HealthCheck)

	// 公开接口
	c.server.Post("/api/register", c.Register)
	c.server.Post("/api/login", c.Login)

	// 需要认证的接口
	c.server.Use(http.MethodGet, "/user/*", c.AuthMiddleware())
	{
		c.server.Get("/user/info/:user_id", c.GetUserInfo)
		c.server.Post("/user/info", c.UpdateUserInfo)
		c.server.Post("/user/password", c.ChangePassword)
	}
}

// HealthCheck 健康检查
func (c *Controller) HealthCheck(ctx *web.Context) {
	_ = ctx.RespJSONOK(dto.Success(map[string]any{
		"status":    "ok",
		"service":   "user-service",
		"timestamp": time.Now().Unix(),
	}))
}

// Register 用户注册
func (c *Controller) Register(ctx *web.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=20"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	user, err := c.userService.Register(ctx.Req.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrUserExists, err.Error()))
		return
	}

	_ = ctx.RespJSONOK(dto.Success(user))
}

// Login 用户登录
func (c *Controller) Login(ctx *web.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	user, err := c.userService.Login(ctx.Req.Context(), req.Username, req.Password)
	if err != nil {
		_ = ctx.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrPasswordInvalid, err.Error()))
		return
	}

	// 生成 JWT 令牌
	token, err := util.GenerateToken(user.ID, user.Role)
	if err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, "生成令牌失败"))
		return
	}

	_ = ctx.RespJSONOK(dto.Success(map[string]any{
		"token": token,
		"user":  user,
	}))
}

// GetUserInfo 获取用户信息
func (c *Controller) GetUserInfo(ctx *web.Context) {
	userID, err := ctx.PathValue("user_id").AsInt64()
	if err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	user, err := c.userService.GetUserInfo(ctx.Req.Context(), userID)
	if err != nil {
		_ = ctx.RespJSON(http.StatusNotFound, dto.Error(errcode.ErrUserNotFound, err.Error()))
		return
	}

	_ = ctx.RespJSONOK(dto.Success(user))
}

// UpdateUserInfo 更新用户信息
func (c *Controller) UpdateUserInfo(ctx *web.Context) {
	userID, err := ctx.PathValue("user_id").AsInt64()
	if err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
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

	if err := c.userService.UpdateUserInfo(ctx.Req.Context(), userID, updates); err != nil {
		_ = ctx.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}

	_ = ctx.RespJSONOK(dto.SuccessNil())
}

// ChangePassword 修改密码
func (c *Controller) ChangePassword(ctx *web.Context) {
	userID, err := ctx.PathValue("user_id").AsInt64()
	if err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
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

	if err := c.userService.ChangePassword(ctx.Req.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		_ = ctx.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrPasswordInvalid, err.Error()))
		return
	}

	_ = ctx.RespJSONOK(dto.SuccessNil())
}

// AuthMiddleware JWT 认证中间件
func (c *Controller) AuthMiddleware() web.Middleware {
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
				_ = ctx.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrTokenInvalid, "无效的认证令牌"))
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
func (c *Controller) Run(addr string) error {
	return c.server.Start(addr)
}
