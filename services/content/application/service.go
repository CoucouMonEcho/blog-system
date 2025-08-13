package application

import (
	"context"
	"time"

	"blog-system/common/pkg/logger"
	"blog-system/services/content/domain"
)

type ContentAppService struct {
	repo   domain.ContentRepository
	logger logger.Logger
}

func NewContentService(repo domain.ContentRepository, lgr logger.Logger) *ContentAppService {
	return &ContentAppService{repo: repo, logger: lgr}
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
