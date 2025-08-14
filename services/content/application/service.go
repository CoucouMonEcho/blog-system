package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"blog-system/common/pkg/logger"
	"blog-system/services/content/domain"

	"github.com/CoucouMonEcho/go-framework/cache"
)

type ContentAppService struct {
	repo   domain.ContentRepository
	logger logger.Logger
	cache  cache.Cache
}

func NewContentService(repo domain.ContentRepository, lgr logger.Logger, c cache.Cache) *ContentAppService {
	return &ContentAppService{repo: repo, logger: lgr, cache: c}
}

func (s *ContentAppService) Create(ctx context.Context, a *domain.Article) (*domain.Article, error) {
	if a.PublishedAt == nil && a.Status == 1 {
		now := time.Now()
		a.PublishedAt = &now
	}
	now := time.Now()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	if err := s.repo.CreateArticle(ctx, a); err != nil {
		s.logger.LogWithContext("content-service", "application", "ERROR", "创建文章失败: %v", err)
		return nil, err
	}
	// 简单回查：以 slug 为唯一键可以添加，暂时省略，调用方可再查
	return a, nil
}

func (s *ContentAppService) GetByID(ctx context.Context, id int64) (*domain.Article, error) {
	return s.repo.GetArticleByID(ctx, id)
}

// Update 更新文章
func (s *ContentAppService) Update(ctx context.Context, a *domain.Article) error {
	if a.ID == 0 {
		return fmt.Errorf("invalid id")
	}
	// 发布状态且未设置发布时间时自动设置
	if a.PublishedAt == nil && a.Status == 1 {
		now := time.Now()
		a.PublishedAt = &now
	}
	a.UpdatedAt = time.Now()
	return s.repo.UpdateArticle(ctx, a)
}

// Delete 删除文章（逻辑删除）
func (s *ContentAppService) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return fmt.Errorf("invalid id")
	}
	return s.repo.DeleteArticle(ctx, id)
}

// ListSummaries 分页查询文章摘要
func (s *ContentAppService) ListSummaries(ctx context.Context, page, pageSize int) ([]*domain.ArticleSummary, int64, error) {
	return s.repo.ListArticleSummaries(ctx, page, pageSize)
}

// SearchSummaries 关键词搜索文章摘要
func (s *ContentAppService) SearchSummaries(ctx context.Context, keyword string, page, pageSize int) ([]*domain.ArticleSummary, int64, error) {
	return s.repo.SearchArticleSummaries(ctx, keyword, page, pageSize)
}

// GetCategoryTree 树状分类，带缓存（简单内存缓存可后续替换为 Redis）
func (s *ContentAppService) GetCategoryTree(ctx context.Context) ([]*domain.CategoryNode, error) {
	if s.cache != nil {
		if v, err := s.cache.Get(ctx, "content:category_tree"); err == nil {
			if str, ok := v.(string); ok {
				var cached []*domain.CategoryNode
				if json.Unmarshal([]byte(str), &cached) == nil {
					return cached, nil
				}
			}
		}
	}
	list, err := s.repo.ListAllCategories(ctx)
	if err != nil {
		return nil, err
	}
	id2node := make(map[int64]*domain.CategoryNode)
	var roots []*domain.CategoryNode
	for _, c := range list {
		id2node[c.ID] = &domain.CategoryNode{ID: c.ID, Name: c.Name, Slug: c.Slug}
	}
	for _, c := range list {
		node := id2node[c.ID]
		if c.ParentID == 0 {
			roots = append(roots, node)
		} else if p, ok := id2node[c.ParentID]; ok {
			p.Children = append(p.Children, node)
		} else {
			roots = append(roots, node)
		}
	}
	if s.cache != nil {
		if b, err := json.Marshal(roots); err == nil {
			_ = s.cache.Set(ctx, "content:category_tree", string(b), 5*time.Minute)
		}
	}
	return roots, nil
}
