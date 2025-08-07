package api

import (
	"blog-system/common/pkg/dto"
	"blog-system/common/pkg/errcode"
	"blog-system/common/pkg/util"
	"blog-system/services/user/application"
	"github.com/CoucouMonEcho/go-framework/web"
	"net/http"
)

// HTTPServer HTTP 服务器
type HTTPServer struct {
	userService *application.UserAppService
	engine      *web.HTTPServer
}

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(userService *application.UserAppService) *HTTPServer {
	//TODO middleware
	engine := web.NewHTTPServer()
	server := &HTTPServer{
		userService: userService,
		engine:      engine,
	}
	server.registerRoutes()
	return server
}

// registerRoutes 注册路由
func (server *HTTPServer) registerRoutes() {
	// 公开接口
	server.engine.Post("/api/register", server.Register)
	server.engine.Post("/api/login", server.Login)

	// 需要认证的接口
	server.engine.Use(http.MethodGet, "/user/*", server.AuthMiddleware())
	{
		server.engine.Get("/user/info/:user_id", server.GetUserInfo)
		server.engine.Post("/user/info/:user_id", server.UpdateUserInfo)
		server.engine.Post("/user/password/:user_id", server.ChangePassword)
	}
}

// Register 用户注册
func (server *HTTPServer) Register(c *web.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=20"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.BindJSON(&req); err != nil {
		_ = c.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	user, err := server.userService.Register(c.Req.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		_ = c.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrUserExists, err.Error()))
		return
	}

	_ = c.RespJSONOK(dto.Success(user))
}

// Login 用户登录
func (server *HTTPServer) Login(c *web.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		_ = c.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	user, err := server.userService.Login(c.Req.Context(), req.Username, req.Password)
	if err != nil {
		_ = c.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrPasswordInvalid, err.Error()))
		return
	}

	// 生成 JWT 令牌
	token, err := util.GenerateToken(user.ID, user.Role)
	if err != nil {
		_ = c.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, "生成令牌失败"))
		return
	}

	_ = c.RespJSONOK(dto.Success(map[string]any{
		"token": token,
		"user":  user,
	}))
}

// GetUserInfo 获取用户信息
func (server *HTTPServer) GetUserInfo(c *web.Context) {
	userID, err := c.PathValue("user_id").AsInt64()
	if err != nil {
		_ = c.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	user, err := server.userService.GetUserInfo(c.Req.Context(), userID)
	if err != nil {
		_ = c.RespJSON(http.StatusNotFound, dto.Error(errcode.ErrUserNotFound, err.Error()))
		return
	}

	_ = c.RespJSONOK(dto.Success(user))
}

// UpdateUserInfo 更新用户信息
func (server *HTTPServer) UpdateUserInfo(c *web.Context) {
	userID, err := c.PathValue("user_id").AsInt64()
	if err != nil {
		_ = c.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	var req struct {
		Username string `json:"username,omitempty"`
		Email    string `json:"email,omitempty"`
		Avatar   string `json:"avatar,omitempty"`
	}

	if err := c.BindJSON(&req); err != nil {
		_ = c.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
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

	if err := server.userService.UpdateUserInfo(c.Req.Context(), userID, updates); err != nil {
		_ = c.RespJSON(http.StatusInternalServerError, dto.Error(errcode.ErrInternal, err.Error()))
		return
	}

	_ = c.RespJSONOK(dto.SuccessNil())
}

// ChangePassword 修改密码
func (server *HTTPServer) ChangePassword(c *web.Context) {
	userID, err := c.PathValue("user_id").AsInt64()
	if err != nil {
		_ = c.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.BindJSON(&req); err != nil {
		_ = c.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrParam, err.Error()))
		return
	}

	if err := server.userService.ChangePassword(c.Req.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		_ = c.RespJSON(http.StatusBadRequest, dto.Error(errcode.ErrPasswordInvalid, err.Error()))
		return
	}

	_ = c.RespJSONOK(dto.SuccessNil())
}

// AuthMiddleware JWT 认证中间件
func (server *HTTPServer) AuthMiddleware() web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(c *web.Context) {
			token := c.Req.Header.Get("Authorization")
			if token == "" {
				_ = c.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrUnauthorized, "缺少认证令牌"))
				return
			}

			// 移除 "Bearer " 前缀
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			claims, err := util.ParseToken(token)
			if err != nil {
				_ = c.RespJSON(http.StatusUnauthorized, dto.Error(errcode.ErrTokenInvalid, "无效的认证令牌"))
				return
			}

			// 存储用户信息
			c.UserValues["user_id"] = claims.UserID
			c.UserValues["user_role"] = claims.Role

			// next
			next(c)
		}
	}
}

// Run 启动服务器
func (server *HTTPServer) Run(addr string) error {
	return server.engine.Start(addr)
}
