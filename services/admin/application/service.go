package application

import (
	"blog-system/common/pkg/logger"
	"blog-system/services/admin/domain"
	"context"
	"errors"
	"time"

	"github.com/CoucouMonEcho/go-framework/cache"
)

type AdminService struct {
	Users   UserClient
	Content ContentClient
	Logger  logger.Logger
	Cache   cache.Cache
	Stat    StatClient
	Prom    PromClient
}

// UserClient 抽象 user-service 能力（登录 + 管理）
type UserClient interface {
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error)
}

// ContentClient 抽象 content-service 能力（文章/分类/标签管理）
type ContentClient interface {
	// 文章
	CreateArticle(ctx context.Context, a *domain.Article) error
	UpdateArticle(ctx context.Context, a *domain.Article) error
	DeleteArticle(ctx context.Context, id int64) error
	ListArticles(ctx context.Context) ([]*domain.Article, int64, error)
	CountArticles(ctx context.Context) (int64, error)
	// 分类（全量）
	CreateCategory(ctx context.Context, c *domain.Category) error
	UpdateCategory(ctx context.Context, c *domain.Category) error
	DeleteCategory(ctx context.Context, id int64) error
	ListCategories(ctx context.Context) ([]*domain.Category, int64, error)
	CountCategories(ctx context.Context) (int64, error)
	// 标签（全量）
	CreateTag(ctx context.Context, t *domain.Tag) error
	UpdateTag(ctx context.Context, t *domain.Tag) error
	DeleteTag(ctx context.Context, id int64) error
	ListTags(ctx context.Context) ([]*domain.Tag, error)
	CountTags(ctx context.Context) (int64, error)
}

func NewAdminService(userCli UserClient, contentCli ContentClient, l logger.Logger, cache cache.Cache, stat StatClient, prom PromClient) *AdminService {
	return &AdminService{Users: userCli, Content: contentCli, Logger: l, Cache: cache, Stat: stat, Prom: prom}
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
	return s.Content.CreateArticle(ctx, a)
}
func (s *AdminService) UpdateArticle(ctx context.Context, a *domain.Article) error {
	a.UpdatedAt = time.Now()
	return s.Content.UpdateArticle(ctx, a)
}
func (s *AdminService) DeleteArticle(ctx context.Context, id int64) error {
	return s.Content.DeleteArticle(ctx, id)
}
func (s *AdminService) ListArticles(ctx context.Context) ([]*domain.Article, int64, error) {
	return s.Content.ListArticles(ctx)
}

// 分类管理
func (s *AdminService) CreateCategory(ctx context.Context, c *domain.Category) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	if err := s.Content.CreateCategory(ctx, c); err != nil {
		return err
	}
	if s.Cache != nil {
		_ = s.Cache.Del(ctx, "content:category_tree")
	}
	return nil
}
func (s *AdminService) UpdateCategory(ctx context.Context, c *domain.Category) error {
	c.UpdatedAt = time.Now()
	if err := s.Content.UpdateCategory(ctx, c); err != nil {
		return err
	}
	if s.Cache != nil {
		_ = s.Cache.Del(ctx, "content:category_tree")
	}
	return nil
}
func (s *AdminService) DeleteCategory(ctx context.Context, id int64) error {
	if err := s.Content.DeleteCategory(ctx, id); err != nil {
		return err
	}
	if s.Cache != nil {
		_ = s.Cache.Del(ctx, "content:category_tree")
	}
	return nil
}
func (s *AdminService) ListCategories(ctx context.Context) ([]*domain.Category, int64, error) {
	return s.Content.ListCategories(ctx)
}

// 标签管理（全量）
func (s *AdminService) CreateTag(ctx context.Context, t *domain.Tag) error {
	return s.Content.CreateTag(ctx, t)
}
func (s *AdminService) UpdateTag(ctx context.Context, t *domain.Tag) error {
	return s.Content.UpdateTag(ctx, t)
}
func (s *AdminService) DeleteTag(ctx context.Context, id int64) error {
	return s.Content.DeleteTag(ctx, id)
}
func (s *AdminService) ListTags(ctx context.Context) ([]*domain.Tag, error) {
	return s.Content.ListTags(ctx)
}
func (s *AdminService) CountTags(ctx context.Context) (int64, error) {
	return s.Content.CountTags(ctx)
}

// Dashboard 概览
func (s *AdminService) Dashboard(ctx context.Context) (map[string]int64, error) {
	if s.Stat == nil {
		return nil, errors.New("stat 客户端未初始化")
	}
	pv, uv, online, err := s.Stat.Overview(ctx)
	if err != nil {
		return nil, err
	}
	artTotal, err := s.Content.CountArticles(ctx)
	if err != nil {
		return nil, err
	}
	catTotal, err := s.Content.CountCategories(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]int64{
		"pv_today":          pv,
		"uv_today":          uv,
		"online_users":      online,
		"article_total":     artTotal,
		"category_total":    catTotal,
		"error_5xx_last_1h": 0,
	}, nil
}

// PVSeries 代理到 stat
func (s *AdminService) PVSeries(ctx context.Context, from, to string, interval string) ([]map[string]int64, error) {
	if s.Stat == nil {
		return nil, errors.New("stat 客户端未初始化")
	}
	return s.Stat.PVSeries(ctx, from, to, interval)
}

// StatClient 抽象 stat-service 能力
type StatClient interface {
	Overview(ctx context.Context) (pvToday, uvToday, onlineUsers int64, err error)
	PVSeries(ctx context.Context, from, to, interval string) ([]map[string]int64, error)
}

// PromClient 查询 Prometheus
type PromClient interface {
	ErrorRate(ctx context.Context, from, to, service string) (float64, error)
	LatencyPercentile(ctx context.Context, from, to, service string) (map[string]float64, error)
	TopEndpoints(ctx context.Context, from, to, service string, topN int) ([]map[string]any, error)
	ActiveUsers(ctx context.Context, from, to string) (int64, error)
}

func (s *AdminService) ErrorRate(ctx context.Context, from, to, service string) (float64, error) {
	if s.Prom == nil {
		return 0, errors.New("prometheus 客户端未初始化")
	}
	return s.Prom.ErrorRate(ctx, from, to, service)
}

func (s *AdminService) LatencyPercentile(ctx context.Context, from, to, service string) (map[string]float64, error) {
	if s.Prom == nil {
		return nil, errors.New("prometheus 客户端未初始化")
	}
	return s.Prom.LatencyPercentile(ctx, from, to, service)
}

func (s *AdminService) TopEndpoints(ctx context.Context, from, to, service string, topN int) ([]map[string]any, error) {
	if s.Prom == nil {
		return nil, errors.New("prometheus 客户端未初始化")
	}
	return s.Prom.TopEndpoints(ctx, from, to, service, topN)
}

func (s *AdminService) ActiveUsers(ctx context.Context, from, to string) (int64, error) {
	if s.Prom == nil {
		return 0, errors.New("prometheus 客户端未初始化")
	}
	return s.Prom.ActiveUsers(ctx, from, to)
}
