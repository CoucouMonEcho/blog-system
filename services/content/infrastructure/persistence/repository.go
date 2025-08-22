package infrastructure

import (
	"context"

	"blog-system/services/content/domain"

	"strings"

	"github.com/CoucouMonEcho/go-framework/orm"
)

type ContentRepository struct {
	db *orm.DB
}

func NewContentRepository(db *orm.DB) *ContentRepository { return &ContentRepository{db: db} }

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
	// 使用 go-framework ORM 的 Raw + AsPredicate 进行 LIKE 查询
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

// ListAllCategories 获取所有分类
func (r *ContentRepository) ListAllCategories(ctx context.Context) ([]*domain.Category, error) {
	list, err := orm.NewSelector[domain.Category](r.db).OrderBy(orm.Asc("Sort")).GetMulti(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// ListCategories 分页
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

// CountCategories 数量
func (r *ContentRepository) CountCategories(ctx context.Context) (int64, error) {
	cnt, err := orm.NewSelector[domain.Category](r.db).Select(orm.Count("Id").As("Id")).Get(ctx)
	if err != nil {
		return 0, err
	}
	return cnt.ID, nil
}

// CreateCategory 新增
func (r *ContentRepository) CreateCategory(ctx context.Context, c *domain.Category) error {
	return orm.NewInserter[domain.Category](r.db).Values(c).Exec(ctx).Err()
}

// UpdateCategory 更新
func (r *ContentRepository) UpdateCategory(ctx context.Context, c *domain.Category) error {
	return orm.NewUpdater[domain.Category](r.db).
		Set(orm.C("Name"), c.Name).
		Set(orm.C("Slug"), c.Slug).
		Set(orm.C("Sort"), c.Sort).
		Set(orm.C("UpdatedAt"), c.UpdatedAt).
		Where(orm.C("Id").Eq(c.ID)).
		Exec(ctx).Err()
}

// DeleteCategory 删除
func (r *ContentRepository) DeleteCategory(ctx context.Context, id int64) error {
	return orm.NewDeleter[domain.Category](r.db).Where(orm.C("Id").Eq(id)).Exec(ctx).Err()
}

// UpdateArticle 更新文章
func (r *ContentRepository) UpdateArticle(ctx context.Context, a *domain.Article) error {
	return orm.NewUpdater[domain.Article](r.db).
		Set(orm.C("Title"), a.Title).
		Set(orm.C("Slug"), a.Slug).
		Set(orm.C("Content"), a.Content).
		Set(orm.C("Summary"), a.Summary).
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
	// 逻辑删除：将状态置为 2(私密/删除)
	return orm.NewUpdater[domain.Article](r.db).
		Set(orm.C("Status"), 2).
		Where(orm.C("Id").Eq(id)).
		Exec(ctx).Err()
}

// CountArticles 数量
func (r *ContentRepository) CountArticles(ctx context.Context) (int64, error) {
	cnt, err := orm.NewSelector[domain.Article](r.db).Select(orm.Count("Id").As("Id")).Get(ctx)
	if err != nil {
		return 0, err
	}
	return cnt.ID, nil
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
