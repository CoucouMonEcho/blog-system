package domain

import (
	"context"
	"time"
)

type Comment struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	ArticleID int64     `json:"article_id"`
	ParentID  int64     `json:"parent_id"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CommentRepository interface {
	Create(ctx context.Context, c *Comment) error
	GetByID(ctx context.Context, id int64) (*Comment, error)
}
