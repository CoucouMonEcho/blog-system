package application

import (
	"context"
	"time"

	"blog-system/common/pkg/logger"
	"blog-system/services/comment/domain"
)

type CommentAppService struct {
	repo   domain.CommentRepository
	logger logger.Logger
}

func NewCommentService(repo domain.CommentRepository, lgr logger.Logger) *CommentAppService {
	return &CommentAppService{repo: repo, logger: lgr}
}

func (s *CommentAppService) Create(ctx context.Context, c *domain.Comment) (*domain.Comment, error) {
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	if err := s.repo.Create(ctx, c); err != nil {
		s.logger.LogWithContext("comment-service", "application", "ERROR", "创建评论失败: %v", err)
		return nil, err
	}
	return c, nil
}

func (s *CommentAppService) GetByID(ctx context.Context, id int64) (*domain.Comment, error) {
	return s.repo.GetByID(ctx, id)
}
