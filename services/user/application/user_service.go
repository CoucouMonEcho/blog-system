package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

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
	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(ctx, username); err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if _, err := s.userRepo.FindByEmail(ctx, email); err == nil {
		return nil, errors.New("邮箱已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
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

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *UserAppService) Login(ctx context.Context, username, password string) (*domain.User, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户状态
	if user.Status != 0 {
		return nil, errors.New("用户已被禁用")
	}

	// 验证密码
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("密码错误")
	}

	return user, nil
}

// GetUserInfo 获取用户信息
func (s *UserAppService) GetUserInfo(ctx context.Context, id int64) (*domain.User, error) {
	// 先从缓存获取
	cacheKey := "user:" + string(rune(id))
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		// JSON反序列化缓存数据
		var user domain.User
		if cachedStr, ok := cached.(string); ok {
			if err := json.Unmarshal([]byte(cachedStr), &user); err == nil {
				return &user, nil
			}
		}
	}

	// 从数据库获取
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
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
func (s *UserAppService) UpdateUserInfo(ctx context.Context, id int64, updates map[string]interface{}) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 更新字段
	for key, value := range updates {
		switch key {
		case "username":
			user.Username = value.(string)
		case "email":
			user.Email = value.(string)
		case "avatar":
			user.Avatar = value.(string)
		}
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// 清除缓存
	cacheKey := "user:" + string(rune(id))
	_ = s.cache.Del(ctx, cacheKey)

	return nil
}

// ChangePassword 修改密码
func (s *UserAppService) ChangePassword(ctx context.Context, id int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
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
		return err
	}

	// TODO: 发送重置密码邮件
	// 这里应该生成重置令牌并发送邮件
	// 可以使用user.Email发送重置邮件
	_ = user // 暂时忽略，后续实现邮件发送功能

	return nil
}
