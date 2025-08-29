package infrastructure

import (
	"context"
	"time"

	"blog-system/services/content/domain"

	"strings"

	"github.com/CoucouMonEcho/go-framework/orm"
)

type ContentRepository struct {
	db *orm.DB
}

func NewContentRepository(db *orm.DB) *ContentRepository { return &ContentRepository{db: db} }

// Article
func (r *ContentRepository) CreateArticle(ctx context.Context, a *domain.Article) error {
	res := orm.NewInserter[domain.Article](r.db).
		Values(a).Exec(ctx)
	return res.Err()
}

func (r *ContentRepository) GetArticleByID(ctx context.Context, id int64) (*domain.Article, error) {
	return orm.NewSelector[domain.Article](r.db).Where(orm.C("Id").Eq(id)).Get(ctx)
}

func (r *ContentRepository) ListArticles(ctx context.Context, page, pageSize int) ([]*domain.Article, int64, error) {
	offset := (page - 1) * pageSize
	list, err := orm.NewSelector[domain.Article](r.db).
		OrderBy(orm.Desc("PublishedAt")).
		Limit(pageSize).Offset(offset).GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	cnt, err := orm.NewSelector[domain.Article](r.db).Select(orm.Count("Id").As("Id")).Get(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, cnt.ID, nil
}

// ListArticleSummaries 仅返回 ID 与 Title
func (r *ContentRepository) ListArticleSummaries(ctx context.Context, page, pageSize int) ([]*domain.ArticleSummary, int64, error) {
	offset := (page - 1) * pageSize
	rows, err := orm.NewSelector[domain.Article](r.db).
		OrderBy(orm.Desc("PublishedAt")).
		Limit(pageSize).Offset(offset).
		GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	summaries := make([]*domain.ArticleSummary, 0, len(rows))
	for _, a := range rows {
		if a == nil {
			continue
		}
		summaries = append(summaries, &domain.ArticleSummary{ID: a.ID, Title: a.Title})
	}
	cnt, err := orm.NewSelector[domain.Article](r.db).Select(orm.Count("Id").As("Id")).Get(ctx)
	if err != nil {
		return nil, 0, err
	}
	return summaries, cnt.ID, nil
}

// SearchArticleSummaries 模糊搜索标题或摘要
func (r *ContentRepository) SearchArticleSummaries(ctx context.Context, keyword string, page, pageSize int) ([]*domain.ArticleSummary, int64, error) {
	offset := (page - 1) * pageSize
	if keyword == "" {
		keyword = "%"
	} else {
		keyword = "%" + keyword + "%"
	}
	rows, err := orm.NewSelector[domain.Article](r.db).
		Where(orm.Raw("Title LIKE ? OR Summary LIKE ?", keyword, keyword).AsPredicate()).
		OrderBy(orm.Desc("PublishedAt")).
		Limit(pageSize).Offset(offset).
		GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	summaries := make([]*domain.ArticleSummary, 0, len(rows))
	for _, a := range rows {
		summaries = append(summaries, &domain.ArticleSummary{ID: a.ID, Title: a.Title})
	}
	all, err := orm.NewSelector[domain.Article](r.db).
		Where(orm.Raw("Title LIKE ? OR Summary LIKE ?", keyword, keyword).AsPredicate()).
		GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	return summaries, int64(len(all)), nil
}

// UpdateArticle 更新文章
func (r *ContentRepository) UpdateArticle(ctx context.Context, a *domain.Article) error {
	return orm.NewUpdater[domain.Article](r.db).
		Set(orm.C("Title"), a.Title).
		Set(orm.C("Slug"), a.Slug).
		Set(orm.C("Content"), a.Content).
		Set(orm.C("Summary"), a.Summary).
		Set(orm.C("Cover"), a.Cover).
		Set(orm.C("CategoryID"), a.CategoryID).
		Set(orm.C("Status"), a.Status).
		Set(orm.C("IsTop"), a.IsTop).
		Set(orm.C("IsRecommend"), a.IsRecommend).
		Set(orm.C("PublishedAt"), a.PublishedAt).
		Set(orm.C("UpdatedAt"), a.UpdatedAt).
		Where(orm.C("Id").Eq(a.ID)).
		Exec(ctx).Err()
}

func (r *ContentRepository) DeleteArticle(ctx context.Context, id int64) error {
	// 物理删除文章，并删除文章标签关联
	if err := orm.NewDeleter[domain.Article](r.db).Where(orm.C("Id").Eq(id)).Exec(ctx).Err(); err != nil {
		return err
	}
	return orm.NewDeleter[domain.ArticleTag](r.db).Where(orm.C("ArticleID").Eq(id)).Exec(ctx).Err()
}

// CountArticles 数量
func (r *ContentRepository) CountArticles(ctx context.Context) (int64, error) {
	cnt, err := orm.NewSelector[domain.Article](r.db).Select(orm.Count("Id").As("Id")).Get(ctx)
	if err != nil {
		return 0, err
	}
	return cnt.ID, nil
}

// Category（单级）
func (r *ContentRepository) ListAllCategories(ctx context.Context) ([]*domain.Category, error) {
	list, err := orm.NewSelector[domain.Category](r.db).OrderBy(orm.Asc("Sort")).GetMulti(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *ContentRepository) ListCategories(ctx context.Context, page, pageSize int) ([]*domain.Category, int64, error) {
	offset := (page - 1) * pageSize
	list, err := orm.NewSelector[domain.Category](r.db).OrderBy(orm.Asc("Sort")).Limit(pageSize).Offset(offset).GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	cnt, err := orm.NewSelector[domain.Category](r.db).Select(orm.Count("Id").As("Id")).Get(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, cnt.ID, nil
}

func (r *ContentRepository) CountCategories(ctx context.Context) (int64, error) {
	cnt, err := orm.NewSelector[domain.Category](r.db).Select(orm.Count("Id").As("Id")).Get(ctx)
	if err != nil {
		return 0, err
	}
	return cnt.ID, nil
}

func (r *ContentRepository) CreateCategory(ctx context.Context, c *domain.Category) error {
	return orm.NewInserter[domain.Category](r.db).Values(c).Exec(ctx).Err()
}

func (r *ContentRepository) UpdateCategory(ctx context.Context, c *domain.Category) error {
	return orm.NewUpdater[domain.Category](r.db).
		Set(orm.C("Name"), c.Name).
		Set(orm.C("Slug"), c.Slug).
		Set(orm.C("Sort"), c.Sort).
		Set(orm.C("UpdatedAt"), c.UpdatedAt).
		Where(orm.C("Id").Eq(c.ID)).
		Exec(ctx).Err()
}

func (r *ContentRepository) DeleteCategory(ctx context.Context, id int64) error {
	return orm.NewDeleter[domain.Category](r.db).Where(orm.C("Id").Eq(id)).Exec(ctx).Err()
}

// Tag
func (r *ContentRepository) CreateTag(ctx context.Context, t *domain.Tag) error {
	return orm.NewInserter[domain.Tag](r.db).Values(t).Exec(ctx).Err()
}

func (r *ContentRepository) UpdateTag(ctx context.Context, t *domain.Tag) error {
	return orm.NewUpdater[domain.Tag](r.db).
		Set(orm.C("Name"), t.Name).
		Set(orm.C("Slug"), t.Slug).
		Set(orm.C("Color"), t.Color).
		Set(orm.C("UpdatedAt"), t.UpdatedAt).
		Where(orm.C("Id").Eq(t.ID)).
		Exec(ctx).Err()
}

func (r *ContentRepository) DeleteTag(ctx context.Context, id int64) error {
	// 删除标签与文章关联，再删除标签
	if err := orm.NewDeleter[domain.ArticleTag](r.db).Where(orm.C("TagID").Eq(id)).Exec(ctx).Err(); err != nil {
		return err
	}
	return orm.NewDeleter[domain.Tag](r.db).Where(orm.C("Id").Eq(id)).Exec(ctx).Err()
}

func (r *ContentRepository) ListTags(ctx context.Context, page, pageSize int) ([]*domain.Tag, int64, error) {
	offset := (page - 1) * pageSize
	list, err := orm.NewSelector[domain.Tag](r.db).OrderBy(orm.Asc("Id")).Limit(pageSize).Offset(offset).GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	cnt, err := orm.NewSelector[domain.Tag](r.db).Select(orm.Count("Id").As("Id")).Get(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, cnt.ID, nil
}

func (r *ContentRepository) CountTags(ctx context.Context) (int64, error) {
	cnt, err := orm.NewSelector[domain.Tag](r.db).Select(orm.Count("Id").As("Id")).Get(ctx)
	if err != nil {
		return 0, err
	}
	return cnt.ID, nil
}

func (r *ContentRepository) ListArticleTags(ctx context.Context, articleID int64) ([]*domain.Tag, error) {
	rows, err := orm.NewSelector[domain.Tag](r.db).
		Where(
			orm.Raw("EXISTS (SELECT 1 FROM blog_article_tags at WHERE at.tag_id = blog_tag.id AND at.article_id = ?)", articleID).AsPredicate(),
		).
		GetMulti(ctx)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ContentRepository) UpdateArticleTags(ctx context.Context, articleID int64, tagIDs []int64) error {
	// 先删除旧关联
	if err := orm.NewDeleter[domain.ArticleTag](r.db).Where(orm.C("ArticleID").Eq(articleID)).Exec(ctx).Err(); err != nil {
		return err
	}
	// 批量插入新关联
	batch := make([]*domain.ArticleTag, 0, len(tagIDs))
	for _, tid := range tagIDs {
		batch = append(batch, &domain.ArticleTag{ArticleID: articleID, TagID: tid, CreatedAt: time.Now()})
	}
	if len(batch) == 0 {
		return nil
	}
	return orm.NewInserter[domain.ArticleTag](r.db).Values(batch...).Exec(ctx).Err()
}

// containsFold 简单不区分大小写包含
func containsFold(s, sub string) bool {
	if sub == "" {
		return true
	}
	return stringContainsFold(s, sub)
}

func stringContainsFold(s, sub string) bool {
	return len(s) >= len(sub) && (len(sub) == 0 || (indexFold(s, sub) >= 0))
}

func indexFold(s, sub string) int {
	// 简化处理：统一转小写
	return index(strings.ToLower(s), strings.ToLower(sub))
}

func index(s, sub string) int { return strings.Index(s, sub) }
