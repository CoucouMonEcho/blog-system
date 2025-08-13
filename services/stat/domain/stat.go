package domain

import (
	"context"
)

type Metric struct {
	ID         int64  `json:"id"`
	Type       string `json:"type"`
	TargetID   int64  `json:"target_id"`
	TargetType string `json:"target_type"`
	UserID     *int64 `json:"user_id"`
	Count      int64  `json:"count"`
}

type StatRepository interface {
	Incr(ctx context.Context, typ string, targetID int64, targetType string, userID *int64) error
	Get(ctx context.Context, typ string, targetID int64, targetType string, userID *int64) (int64, error)
}
