package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"github.com/CoucouMonEcho/go-framework/orm"
)

type StatRepository struct{ db *orm.DB }

func NewStatRepository(db *orm.DB) *StatRepository { return &StatRepository{db: db} }

func (r *StatRepository) Incr(ctx context.Context, typ string, targetID int64, targetType string, userID *int64) error {
	// 简化：直接使用 upsert 或分两步
	// 这里用 ORM Upsert 如果支持，否则先查再插/更
	return errors.New("not implemented: use raw SQL with upsert if available in framework")
}

func (r *StatRepository) Get(ctx context.Context, typ string, targetID int64, targetType string, userID *int64) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
