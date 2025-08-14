package application

import (
	"blog-system/common/pkg/logger"
	"blog-system/services/stat/domain"
	"context"
)

type StatAppService struct {
	repo   domain.StatRepository
	logger logger.Logger
}

func NewStatService(repo domain.StatRepository, lgr logger.Logger) *StatAppService {
	return &StatAppService{repo: repo, logger: lgr}
}

func (s *StatAppService) Incr(ctx context.Context, typ string, targetID int64, targetType string, userID *int64) error {
	return s.repo.Incr(ctx, typ, targetID, targetType, userID)
}

func (s *StatAppService) Get(ctx context.Context, typ string, targetID int64, targetType string, userID *int64) (int64, error) {
	return s.repo.Get(ctx, typ, targetID, targetType, userID)
}
