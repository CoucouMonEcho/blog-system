package application

import (
	"blog-system/common/pkg/logger"
	"blog-system/services/admin/domain"
	"context"
	"time"

	"github.com/CoucouMonEcho/go-framework/cache"
)

type AdminService struct {
	Users      domain.UserRepository
	Articles   domain.ArticleRepository
	Categories domain.CategoryRepository
	Logger     logger.Logger
	Cache      cache.Cache
}

func NewAdminService(u domain.UserRepository, a domain.ArticleRepository, c domain.CategoryRepository, l logger.Logger, cache cache.Cache) *AdminService {
	return &AdminService{Users: u, Articles: a, Categories: c, Logger: l, Cache: cache}
}

// 用户管理
func (s *AdminService) CreateUser(ctx context.Context, u *domain.User) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	return s.Users.Create(ctx, u)
}
func (s *AdminService) UpdateUser(ctx context.Context, u *domain.User) error {
	u.UpdatedAt = time.Now()
	return s.Users.Update(ctx, u)
}
func (s *AdminService) DeleteUser(ctx context.Context, id int64) error {
	return s.Users.Delete(ctx, id)
}
func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	return s.Users.List(ctx, page, pageSize)
}

// 文章管理
func (s *AdminService) CreateArticle(ctx context.Context, a *domain.Article) error {
	now := time.Now()
	if a.PublishedAt == nil && a.Status == 1 {
		a.PublishedAt = &now
	}
	a.CreatedAt = now
	a.UpdatedAt = now
	return s.Articles.Create(ctx, a)
}
func (s *AdminService) UpdateArticle(ctx context.Context, a *domain.Article) error {
	a.UpdatedAt = time.Now()
	return s.Articles.Update(ctx, a)
}
func (s *AdminService) DeleteArticle(ctx context.Context, id int64) error {
	return s.Articles.Delete(ctx, id)
}
func (s *AdminService) ListArticles(ctx context.Context, page, pageSize int) ([]*domain.Article, int64, error) {
	return s.Articles.List(ctx, page, pageSize)
}

// 分类管理
func (s *AdminService) CreateCategory(ctx context.Context, c *domain.Category) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	if err := s.Categories.Create(ctx, c); err != nil {
		return err
	}
	if s.Cache != nil {
		_ = s.Cache.Del(ctx, "content:category_tree")
	}
	return nil
}
func (s *AdminService) UpdateCategory(ctx context.Context, c *domain.Category) error {
	c.UpdatedAt = time.Now()
	if err := s.Categories.Update(ctx, c); err != nil {
		return err
	}
	if s.Cache != nil {
		_ = s.Cache.Del(ctx, "content:category_tree")
	}
	return nil
}
func (s *AdminService) DeleteCategory(ctx context.Context, id int64) error {
	if err := s.Categories.Delete(ctx, id); err != nil {
		return err
	}
	if s.Cache != nil {
		_ = s.Cache.Del(ctx, "content:category_tree")
	}
	return nil
}
func (s *AdminService) ListCategories(ctx context.Context, page, pageSize int) ([]*domain.Category, int64, error) {
	return s.Categories.List(ctx, page, pageSize)
}
