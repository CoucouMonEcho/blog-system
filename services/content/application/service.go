package application

import (
	"context"
	"fmt"
	"math"
	"time"

	"blog-system/common/pkg/logger"
	"blog-system/services/content/domain"

	"github.com/CoucouMonEcho/go-framework/cache"
)

// ContentAppService 内容应用服务
type ContentAppService struct {
	repo   domain.ContentRepository
	logger logger.Logger
	cache  cache.Cache
}

func NewContentService(repo domain.ContentRepository, lgr logger.Logger, c cache.Cache) *ContentAppService {
	return &ContentAppService{repo: repo, logger: lgr, cache: c}
}

// 文章相关
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
		s.logger.Error("application: 创建文章失败: %v", err)
		return nil, err
	}
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
	if a.PublishedAt == nil && a.Status == 1 {
		now := time.Now()
		a.PublishedAt = &now
	}
	a.UpdatedAt = time.Now()
	return s.repo.UpdateArticle(ctx, a)
}

// Delete 删除文章（物理删除 + 关联删除在仓储实现）
func (s *ContentAppService) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return fmt.Errorf("invalid id")
	}
	return s.repo.DeleteArticle(ctx, id)
}

// ListSummaries 分页查询文章摘要（内部复用）
func (s *ContentAppService) ListSummaries(ctx context.Context, page, pageSize int) ([]*domain.ArticleSummary, int64, error) {
	return s.repo.ListArticleSummaries(ctx, page, pageSize)
}

// ListSummariesFiltered 支持分类与标签过滤的文章摘要列表
func (s *ContentAppService) ListSummariesFiltered(ctx context.Context, categoryID *int64, tagIDs []int64, page, pageSize int) ([]*domain.ArticleSummary, int64, error) {
	// 优先调用有过滤实现的方法；若仓储未实现则回退到无过滤，并在上层过滤（此处假定已有实现）
	type repoWithFilter interface {
		ListArticleSummariesFiltered(ctx context.Context, categoryID *int64, tagIDs []int64, page, pageSize int) ([]*domain.ArticleSummary, int64, error)
	}
	if rf, ok := s.repo.(repoWithFilter); ok {
		return rf.ListArticleSummariesFiltered(ctx, categoryID, tagIDs, page, pageSize)
	}
	return s.repo.ListArticleSummaries(ctx, page, pageSize)
}

// SearchSummaries 关键词搜索文章摘要
func (s *ContentAppService) SearchSummaries(ctx context.Context, keyword string, page, pageSize int) ([]*domain.ArticleSummary, int64, error) {
	return s.repo.SearchArticleSummaries(ctx, keyword, page, pageSize)
}

// 全量文章列表与计数（用于对外接口不分页）
func (s *ContentAppService) CountArticles(ctx context.Context) (int64, error) {
	return s.repo.CountArticles(ctx)
}

func (s *ContentAppService) ListAllArticles(ctx context.Context) ([]*domain.Article, int64, error) {
	// 复用分页接口：一次拉取全部
	const m = int(^uint(0) >> 1)
	list, total, err := s.repo.ListArticles(ctx, 1, m)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// 分类（单级）
func (s *ContentAppService) ListCategories(ctx context.Context, page, pageSize int) ([]*domain.Category, int64, error) {
	return s.repo.ListCategories(ctx, page, pageSize)
}

func (s *ContentAppService) ListAllCategories(ctx context.Context) ([]*domain.Category, error) {
	return s.repo.ListAllCategories(ctx)
}

func (s *ContentAppService) CountCategories(ctx context.Context) (int64, error) {
	return s.repo.CountCategories(ctx)
}

func (s *ContentAppService) UpdateCategory(ctx context.Context, c *domain.Category) error {
	c.UpdatedAt = time.Now()
	return s.repo.UpdateCategory(ctx, c)
}

func (s *ContentAppService) DeleteCategory(ctx context.Context, id int64) error {
	return s.repo.DeleteCategory(ctx, id)
}

func (s *ContentAppService) CreateCategory(ctx context.Context, c *domain.Category) error {
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	return s.repo.CreateCategory(ctx, c)
}

// 标签
func (s *ContentAppService) CreateTag(ctx context.Context, t *domain.Tag) error {
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	return s.repo.CreateTag(ctx, t)
}

func (s *ContentAppService) UpdateTag(ctx context.Context, t *domain.Tag) error {
	t.UpdatedAt = time.Now()
	return s.repo.UpdateTag(ctx, t)
}

func (s *ContentAppService) DeleteTag(ctx context.Context, id int64) error {
	return s.repo.DeleteTag(ctx, id)
}

func (s *ContentAppService) ListTags(ctx context.Context, page, pageSize int) ([]*domain.Tag, int64, error) {
	return s.repo.ListTags(ctx, page, pageSize)
}

func (s *ContentAppService) CountTags(ctx context.Context) (int64, error) {
	return s.repo.CountTags(ctx)
}

func (s *ContentAppService) ListAllTags(ctx context.Context) ([]*domain.Tag, int64, error) {
	list, total, err := s.repo.ListTags(ctx, 1, math.MaxInt)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *ContentAppService) ListAllTagsWithCount(ctx context.Context) ([]struct {
	Tag   *domain.Tag
	Count int64
}, error) {
	all, err := s.repo.ListAllTags(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]struct {
		Tag   *domain.Tag
		Count int64
	}, 0, len(all))
	for _, t := range all {
		c, err := s.repo.CountArticlesByTag(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		res = append(res, struct {
			Tag   *domain.Tag
			Count int64
		}{Tag: t, Count: c})
	}
	return res, nil
}

func (s *ContentAppService) GetArticleTags(ctx context.Context, articleID int64) ([]*domain.Tag, error) {
	return s.repo.ListArticleTags(ctx, articleID)
}

func (s *ContentAppService) SetArticleTags(ctx context.Context, articleID int64, tagIDs []int64) error {
	return s.repo.UpdateArticleTags(ctx, articleID, tagIDs)
}
