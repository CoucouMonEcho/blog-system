package domain

import (
	"context"
	"time"
)

type Article struct {
	ID           int64      `json:"id"`
	Title        string     `json:"title"`
	Slug         string     `json:"slug"`
	Content      string     `json:"content"`
	Summary      string     `json:"summary"`
	AuthorID     int64      `json:"author_id"`
	CategoryID   int64      `json:"category_id"`
	Status       int        `json:"status"`
	ViewCount    int64      `json:"view_count"`
	LikeCount    int64      `json:"like_count"`
	CommentCount int64      `json:"comment_count"`
	IsTop        bool       `json:"is_top"`
	IsRecommend  bool       `json:"is_recommend"`
	PublishedAt  *time.Time `json:"published_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ContentRepository interface {
	CreateArticle(ctx context.Context, a *Article) error
	GetArticleByID(ctx context.Context, id int64) (*Article, error)
	ListArticles(ctx context.Context, page, pageSize int) ([]*Article, int64, error)
}
