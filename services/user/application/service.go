package application

import (
	"blog-system/common/pkg/util"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"blog-system/common/pkg/logger"
	"blog-system/services/user/domain"

	"github.com/CoucouMonEcho/go-framework/cache"
	"golang.org/x/crypto/bcrypt"
)

// UserAppService 用户应用服务
type UserAppService struct {
	userRepo domain.UserRepository
	cache    cache.Cache
}

// NewUserService 创建用户服务
func NewUserService(userRepo domain.UserRepository, cache cache.Cache) *UserAppService {
	return &UserAppService{
		userRepo: userRepo,
		cache:    cache,
	}
}

// Register 用户注册
func (s *UserAppService) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	if _, err := s.userRepo.FindByUsername(ctx, username); err == nil {
		return nil, errors.New("用户名已存在")
	}
	if _, err := s.userRepo.FindByEmail(ctx, email); err == nil {
		return nil, errors.New("邮箱已存在")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log().Error("application: 密码加密失败: %v", err)
		return nil, err
	}
	now := time.Now()
	user := &domain.User{
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		Role:      "user",
		Status:    0, // 正常状态
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err = s.userRepo.Create(ctx, user); err != nil {
		logger.Log().Error("application: 用户创建失败: %v", err)
		return nil, err
	}
	logger.Log().Info("application: 注册成功: username=%s", user.Username)
	return user, nil
}

// Login 用户登录
func (s *UserAppService) Login(ctx context.Context, username, password string) (*domain.User, string, error) {
	uname := strings.TrimSpace(username)
	if uname == "" {
		return nil, "", errors.New("用户名不能为空")
	}
	user, err := s.userRepo.FindByUsername(ctx, uname)
	if err != nil {
		logger.Log().Warn("application: 登录失败: 用户不存在, username=%s, err=%v", uname, err)
		return nil, "", errors.New("用户不存在")
	}
	// 检查用户状态
	if user.Status != 0 {
		return nil, "", errors.New("用户已被禁用")
	}
	// 验证密码
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logger.Log().Warn("application: 登录失败: 密码错误, username=%s, err=%v", username, err)
		return nil, "", errors.New("密码错误")
	}
	// 生成 JWT 令牌
	token, err := util.GenerateToken(user.ID, user.Role)
	if err != nil {
		logger.Log().Error("application: 生成token失败: id=%d username=%s err=%v", user.ID, user.Username, err)
		return nil, "", err
	}
	// 将 token 写入缓存，值为用户JSON
	if userData, er := json.Marshal(user); er == nil {
		_ = s.cache.Set(ctx, "token_"+token, string(userData), 24*time.Hour)
	}
	logger.Log().Info("application: 登录成功: id=%d username=%s", user.ID, user.Username)
	return user, token, nil
}

// GetUserInfo 获取用户信息
func (s *UserAppService) GetUserInfo(ctx context.Context, id int64) (*domain.User, error) {
	// 先从缓存获取
	cacheKey := "user_" + strconv.FormatInt(id, 10)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		// JSON反序列化缓存数据
		var user domain.User
		if cachedStr, ok := cached.(string); ok {
			if err := json.Unmarshal([]byte(cachedStr), &user); err == nil {
				logger.Log().Debug("application: 命中缓存: id=%d", id)
				return &user, nil
			}
		}
	}
	// 从数据库获取
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		logger.Log().Error("application: 查询用户失败: id=%d, err=%v", id, err)
		return nil, err
	}
	// 缓存用户信息 - JSON序列化并缓存用户数据
	if userData, err := json.Marshal(user); err == nil {
		// 设置缓存，过期时间30分钟
		_ = s.cache.Set(ctx, cacheKey, string(userData), 30*time.Minute)
	}
	return user, nil
}

// UpdateUserInfo 更新用户信息
func (s *UserAppService) UpdateUserInfo(ctx context.Context, id int64, updates map[string]any) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		logger.Log().Error("application: 更新失败: 读取用户错误 id=%d err=%v", id, err)
		return err
	}
	// 更新字段
	for key, value := range updates {
		switch key {
		case "username":
			if v, ok := value.(string); ok {
				user.Username = v
			}
		case "email":
			if v, ok := value.(string); ok {
				user.Email = v
			}
		case "avatar":
			if v, ok := value.(string); ok {
				user.Avatar = &sql.NullString{String: v, Valid: v != ""}
			}
		case "role":
			if v, ok := value.(string); ok {
				user.Role = v
			}
		}
	}
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Log().Error("application: 更新失败: 写入用户错误 id=%d err=%v", id, err)
		return err
	}
	// 清除缓存
	cacheKey := "user_" + strconv.FormatInt(id, 10)
	_ = s.cache.Del(ctx, cacheKey)
	logger.Log().Info("application: 更新成功: id=%d", id)
	return nil
}

// ChangePassword 修改密码
func (s *UserAppService) ChangePassword(ctx context.Context, id int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		logger.Log().Error("application: 修改密码失败: 读取用户错误 id=%d err=%v", id, err)
		return err
	}
	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		logger.Log().Warn("application: 修改密码失败: 旧密码错误 id=%d, err=%v", id, err)
		return errors.New("旧密码错误")
	}
	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Log().Error("application: 修改密码失败: 加密错误 id=%d err=%v", id, err)
		return err
	}
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

// ResetPassword 重置密码
func (s *UserAppService) ResetPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		logger.Log().Error("application: 重置密码失败: 读取用户错误 email=%s err=%v", email, err)
		return err
	}
	// TODO: 发送重置密码邮件
	_ = user
	return nil
}

// ListUsers 分页列表
func (s *UserAppService) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	return s.userRepo.List(ctx, page, pageSize)
}

// ChangeUserStatus 更新用户状态
func (s *UserAppService) ChangeUserStatus(ctx context.Context, id int64, status int) error {
	return s.userRepo.UpdateStatus(ctx, id, status)
}
