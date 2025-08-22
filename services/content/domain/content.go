package domain

import (
	"context"
	"time"
)

type Article struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Content     string     `json:"content"`
	Summary     string     `json:"summary"`
	AuthorID    int64      `json:"author_id"`
	CategoryID  int64      `json:"category_id"`
	Status      int        `json:"status"`
	ViewCount   int64      `json:"view_count"`
	LikeCount   int64      `json:"like_count"`
	IsTop       bool       `json:"is_top"`
	IsRecommend bool       `json:"is_recommend"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName 显式指定表名，避免命名策略差异
func (Article) TableName() string { return "blog_article" }

// ArticleSummary 文章摘要
type ArticleSummary struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// Category 分类领域模型
type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Sort      int       `json:"sort"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Category) TableName() string { return "blog_category" }

// CategoryNode 分类树节点
type CategoryNode struct {
	ID       int64           `json:"id"`
	Name     string          `json:"name"`
	Slug     string          `json:"slug"`
	Children []*CategoryNode `json:"children,omitempty"`
}

type ContentRepository interface {
	CreateArticle(ctx context.Context, a *Article) error
	GetArticleByID(ctx context.Context, id int64) (*Article, error)
	ListArticles(ctx context.Context, page, pageSize int) ([]*Article, int64, error)
	CountArticles(ctx context.Context) (int64, error)
	UpdateArticle(ctx context.Context, a *Article) error
	DeleteArticle(ctx context.Context, id int64) error
	ListArticleSummaries(ctx context.Context, page, pageSize int) ([]*ArticleSummary, int64, error)
	SearchArticleSummaries(ctx context.Context, keyword string, page, pageSize int) ([]*ArticleSummary, int64, error)

	// Category
	ListAllCategories(ctx context.Context) ([]*Category, error)
	ListCategories(ctx context.Context, page, pageSize int) ([]*Category, int64, error)
	CountCategories(ctx context.Context) (int64, error)
	CreateCategory(ctx context.Context, c *Category) error
	UpdateCategory(ctx context.Context, c *Category) error
	DeleteCategory(ctx context.Context, id int64) error
}
