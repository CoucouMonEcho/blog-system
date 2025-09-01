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
	Cover       string     `json:"cover"`
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

func (Article) TableName() string { return "blog_article" }

// ArticleSummary 文章摘要
type ArticleSummary struct {
	ID       int64          `json:"id"`
	Title    string         `json:"title"`
	Summary  string         `json:"summary,omitempty"`
	AuthorID int64          `json:"author_id,omitempty"`
	Category *CategoryBrief `json:"category,omitempty"`
	Tags     []*TagBrief    `json:"tags,omitempty"`
	CoverURL string         `json:"cover_url,omitempty"`
}

// Category 分类领域模型（单级）
type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Sort      int       `json:"sort"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Category) TableName() string { return "blog_category" }

// 简要返回用
type CategoryBrief struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Tag 标签领域模型
type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Tag) TableName() string { return "blog_tag" }

// 简要返回用
type TagBrief struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Color string `json:"color"`
}

// ArticleTag 文章-标签 关联（多对多）
type ArticleTag struct {
	ID        int64     `json:"id"`
	ArticleID int64     `json:"article_id"`
	TagID     int64     `json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (ArticleTag) TableName() string { return "blog_article_tags" }

// ContentRepository 聚合仓储接口
type ContentRepository interface {

	// Article 文章
	CreateArticle(ctx context.Context, a *Article) error
	GetArticleByID(ctx context.Context, id int64) (*Article, error)
	ListArticles(ctx context.Context, page, pageSize int) ([]*Article, int64, error)
	CountArticles(ctx context.Context) (int64, error)
	UpdateArticle(ctx context.Context, a *Article) error
	DeleteArticle(ctx context.Context, id int64) error
	ListArticleSummaries(ctx context.Context, page, pageSize int) ([]*ArticleSummary, int64, error)
	SearchArticleSummaries(ctx context.Context, keyword string, page, pageSize int) ([]*ArticleSummary, int64, error)

	// Category 分类
	ListAllCategories(ctx context.Context) ([]*Category, error)
	ListCategories(ctx context.Context, page, pageSize int) ([]*Category, int64, error)
	CountCategories(ctx context.Context) (int64, error)
	CreateCategory(ctx context.Context, c *Category) error
	UpdateCategory(ctx context.Context, c *Category) error
	DeleteCategory(ctx context.Context, id int64) error

	// Tag 标签
	CreateTag(ctx context.Context, t *Tag) error
	UpdateTag(ctx context.Context, t *Tag) error
	DeleteTag(ctx context.Context, id int64) error
	ListTags(ctx context.Context, page, pageSize int) ([]*Tag, int64, error)
	CountTags(ctx context.Context) (int64, error)
	ListArticleTags(ctx context.Context, articleID int64) ([]*Tag, error)
	UpdateArticleTags(ctx context.Context, articleID int64, tagIDs []int64) error
	// 新增：标签全量与计数
	ListAllTags(ctx context.Context) ([]*Tag, error)
	CountArticlesByTag(ctx context.Context, tagID int64) (int64, error)
	// 可选：带过滤的摘要
	ListArticleSummariesFiltered(ctx context.Context, categoryID *int64, tagIDs []int64, page, pageSize int) ([]*ArticleSummary, int64, error)
}
