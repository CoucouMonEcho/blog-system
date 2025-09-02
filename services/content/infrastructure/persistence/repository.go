package infrastructure

import (
	"context"
	"time"

	"blog-system/common/pkg/aggregate"
	"blog-system/common/pkg/logger"
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
	return orm.NewSelector[domain.Article](r.db).Where(orm.C("ID").Eq(id)).Get(ctx)
}

func (r *ContentRepository) ListArticles(ctx context.Context, page, pageSize int) ([]*domain.Article, int64, error) {
	offset := (page - 1) * pageSize
	list, err := orm.NewSelector[domain.Article](r.db).
		OrderBy(orm.Desc("PublishedAt")).
		Limit(pageSize).Offset(offset).GetMulti(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListArticles 查询失败: %v", err)
		return nil, 0, err
	}
	cnt, err := orm.NewSelector[aggregate.Result](r.db).From(orm.TableOf(&domain.Article{})).Select(orm.Count("ID").As("Count")).Get(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListArticles 统计失败: %v", err)
		return nil, 0, err
	}
	return list, cnt.Count, nil
}

// ListArticleSummaries 仅返回 ID 与 Title（扩展：填充 summary/author/category/tags/cover_url）
func (r *ContentRepository) ListArticleSummaries(ctx context.Context, page, pageSize int) ([]*domain.ArticleSummary, int64, error) {
	offset := (page - 1) * pageSize
	rows, err := orm.NewSelector[domain.Article](r.db).
		OrderBy(orm.Desc("PublishedAt")).
		Limit(pageSize).Offset(offset).
		GetMulti(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListArticleSummaries 查询失败: %v", err)
		return nil, 0, err
	}
	summaries := make([]*domain.ArticleSummary, 0, len(rows))
	for _, a := range rows {
		if a == nil {
			continue
		}
		s := &domain.ArticleSummary{ID: a.ID, Title: a.Title, AuthorID: a.AuthorID}
		// summary：若空则取 content 前 120 字符（近似）
		if a.Summary != nil && a.Summary.Valid && a.Summary.String != "" {
			s.Summary = a.Summary.String
		} else {
			runes := []rune(a.Content)
			if len(runes) > 120 {
				s.Summary = string(runes[:120])
			} else {
				s.Summary = string(runes)
			}
		}
		// cover_url：取 Cover 字段
		if a.Cover != nil && a.Cover.Valid {
			s.CoverURL = a.Cover.String
		}
		// category 简要
		if a.CategoryID > 0 {
			c, er := orm.NewSelector[domain.Category](r.db).Where(orm.C("ID").Eq(a.CategoryID)).Get(ctx)
			if er == nil && c != nil {
				s.Category = &domain.CategoryBrief{ID: c.ID, Name: c.Name, Slug: c.Slug}
			}
		}
		// tags 简要
		ts, er2 := r.ListArticleTags(ctx, a.ID)
		if er2 == nil && len(ts) > 0 {
			for _, t := range ts {
				color := ""
				if t.Color != nil && t.Color.Valid {
					color = t.Color.String
				}
				s.Tags = append(s.Tags, &domain.TagBrief{ID: t.ID, Name: t.Name, Slug: t.Slug, Color: color})
			}
		}
		summaries = append(summaries, s)
	}
	cnt, err := orm.NewSelector[aggregate.Result](r.db).From(orm.TableOf(&domain.Article{})).Select(orm.Count("ID").As("Count")).Get(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListArticleSummaries 统计失败: %v", err)
		return nil, 0, err
	}
	return summaries, cnt.Count, nil
}

// SearchArticleSummaries 模糊搜索标题或摘要（扩展同上）
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
		logger.Log().Error("infrastructure: SearchArticleSummaries 查询失败: %v", err)
		return nil, 0, err
	}
	summaries := make([]*domain.ArticleSummary, 0, len(rows))
	for _, a := range rows {
		if a == nil {
			continue
		}
		s := &domain.ArticleSummary{ID: a.ID, Title: a.Title, AuthorID: a.AuthorID}
		if a.Summary != nil && a.Summary.Valid && a.Summary.String != "" {
			s.Summary = a.Summary.String
		} else {
			runes := []rune(a.Content)
			if len(runes) > 120 {
				s.Summary = string(runes[:120])
			} else {
				s.Summary = string(runes)
			}
		}
		if a.Cover != nil && a.Cover.Valid {
			s.CoverURL = a.Cover.String
		}
		if a.CategoryID > 0 {
			c, er := orm.NewSelector[domain.Category](r.db).Where(orm.C("ID").Eq(a.CategoryID)).Get(ctx)
			if er == nil && c != nil {
				s.Category = &domain.CategoryBrief{ID: c.ID, Name: c.Name, Slug: c.Slug}
			}
		}
		ts, er2 := r.ListArticleTags(ctx, a.ID)
		if er2 == nil && len(ts) > 0 {
			for _, t := range ts {
				color := ""
				if t.Color != nil && t.Color.Valid {
					color = t.Color.String
				}
				s.Tags = append(s.Tags, &domain.TagBrief{ID: t.ID, Name: t.Name, Slug: t.Slug, Color: color})
			}
		}
		summaries = append(summaries, s)
	}
	allCnt, err := orm.NewSelector[aggregate.Result](r.db).From(orm.TableOf(&domain.Article{})).
		Where(orm.Raw("Title LIKE ? OR Summary LIKE ?", keyword, keyword).AsPredicate()).
		Select(orm.Count("ID").As("Count")).Get(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: SearchArticleSummaries 获取总数失败: %v", err)
		return nil, 0, err
	}
	return summaries, allCnt.Count, nil
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
		Set(orm.C("MetaTitle"), a.MetaTitle).
		Set(orm.C("MetaDesc"), a.MetaDesc).
		Set(orm.C("MetaKeywords"), a.MetaKeywords).
		Set(orm.C("PublishedAt"), a.PublishedAt).
		Set(orm.C("UpdatedAt"), a.UpdatedAt).
		Where(orm.C("ID").Eq(a.ID)).
		Exec(ctx).Err()
}

func (r *ContentRepository) DeleteArticle(ctx context.Context, id int64) error {
	// 物理删除文章，并删除文章标签关联
	if err := orm.NewDeleter[domain.Article](r.db).Where(orm.C("ID").Eq(id)).Exec(ctx).Err(); err != nil {
		logger.Log().Error("infrastructure: DeleteArticle 删除文章失败: %v", err)
		return err
	}
	return orm.NewDeleter[domain.ArticleTag](r.db).Where(orm.C("ArticleID").Eq(id)).Exec(ctx).Err()
}

// CountArticles 数量
func (r *ContentRepository) CountArticles(ctx context.Context) (int64, error) {
	cnt, err := orm.NewSelector[aggregate.Result](r.db).From(orm.TableOf(&domain.Article{})).Select(orm.Count("ID").As("Count")).Get(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: CountArticles 统计失败: %v", err)
		return 0, err
	}
	return cnt.Count, nil
}

// Category（单级）
func (r *ContentRepository) ListAllCategories(ctx context.Context) ([]*domain.Category, error) {
	list, err := orm.NewSelector[domain.Category](r.db).OrderBy(orm.Asc("Sort")).GetMulti(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListAllCategories 查询失败: %v", err)
		return nil, err
	}
	return list, nil
}

func (r *ContentRepository) ListCategories(ctx context.Context, page, pageSize int) ([]*domain.Category, int64, error) {
	offset := (page - 1) * pageSize
	list, err := orm.NewSelector[domain.Category](r.db).OrderBy(orm.Asc("Sort")).Limit(pageSize).Offset(offset).GetMulti(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListCategories 查询失败: %v", err)
		return nil, 0, err
	}
	cnt, err := orm.NewSelector[aggregate.Result](r.db).From(orm.TableOf(&domain.Category{})).Select(orm.Count("ID").As("Count")).Get(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListCategories 统计失败: %v", err)
		return nil, 0, err
	}
	return list, cnt.Count, nil
}

func (r *ContentRepository) CountCategories(ctx context.Context) (int64, error) {
	cnt, err := orm.NewSelector[aggregate.Result](r.db).From(orm.TableOf(&domain.Category{})).Select(orm.Count("ID").As("Count")).Get(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: CountCategories 统计失败: %v", err)
		return 0, err
	}
	return cnt.Count, nil
}

func (r *ContentRepository) CreateCategory(ctx context.Context, c *domain.Category) error {
	return orm.NewInserter[domain.Category](r.db).Values(c).Exec(ctx).Err()
}

func (r *ContentRepository) UpdateCategory(ctx context.Context, c *domain.Category) error {
	return orm.NewUpdater[domain.Category](r.db).
		Set(orm.C("Name"), c.Name).
		Set(orm.C("Slug"), c.Slug).
		Set(orm.C("Description"), c.Description).
		Set(orm.C("Sort"), c.Sort).
		Set(orm.C("UpdatedAt"), c.UpdatedAt).
		Where(orm.C("ID").Eq(c.ID)).
		Exec(ctx).Err()
}

func (r *ContentRepository) DeleteCategory(ctx context.Context, id int64) error {
	return orm.NewDeleter[domain.Category](r.db).Where(orm.C("ID").Eq(id)).Exec(ctx).Err()
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
		Where(orm.C("ID").Eq(t.ID)).
		Exec(ctx).Err()
}

func (r *ContentRepository) DeleteTag(ctx context.Context, id int64) error {
	// 删除标签与文章关联，再删除标签
	if err := orm.NewDeleter[domain.ArticleTag](r.db).Where(orm.C("TagID").Eq(id)).Exec(ctx).Err(); err != nil {
		logger.Log().Error("infrastructure: DeleteTag 删除关联失败: %v", err)
		return err
	}
	return orm.NewDeleter[domain.Tag](r.db).Where(orm.C("ID").Eq(id)).Exec(ctx).Err()
}

func (r *ContentRepository) ListTags(ctx context.Context, page, pageSize int) ([]*domain.Tag, int64, error) {
	offset := (page - 1) * pageSize
	list, err := orm.NewSelector[domain.Tag](r.db).OrderBy(orm.Asc("ID")).Limit(pageSize).Offset(offset).GetMulti(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListTags 查询失败: %v", err)
		return nil, 0, err
	}
	cnt, err := orm.NewSelector[aggregate.Result](r.db).From(orm.TableOf(&domain.Tag{})).Select(orm.Count("ID").As("Count")).Get(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListTags 统计失败: %v", err)
		return nil, 0, err
	}
	return list, cnt.Count, nil
}

func (r *ContentRepository) CountTags(ctx context.Context) (int64, error) {
	cnt, err := orm.NewSelector[aggregate.Result](r.db).From(orm.TableOf(&domain.Tag{})).Select(orm.Count("ID").As("Count")).Get(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: CountTags 统计失败: %v", err)
		return 0, err
	}
	return cnt.Count, nil
}

func (r *ContentRepository) ListArticleTags(ctx context.Context, articleID int64) ([]*domain.Tag, error) {
	rows, err := orm.NewSelector[domain.Tag](r.db).
		Where(
			orm.Raw("EXISTS (SELECT 1 FROM blog_article_tags at WHERE at.tag_id = blog_tag.id AND at.article_id = ?)", articleID).AsPredicate(),
		).
		GetMulti(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListArticleTags 查询失败: %v", err)
		return nil, err
	}
	return rows, nil
}

func (r *ContentRepository) UpdateArticleTags(ctx context.Context, articleID int64, tagIDs []int64) error {
	// 先删除旧关联
	if err := orm.NewDeleter[domain.ArticleTag](r.db).Where(orm.C("ArticleID").Eq(articleID)).Exec(ctx).Err(); err != nil {
		logger.Log().Error("infrastructure: UpdateArticleTags 删除旧关联失败: %v", err)
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

// ListArticleSummariesFiltered 支持按分类与标签过滤的摘要列表
func (r *ContentRepository) ListArticleSummariesFiltered(ctx context.Context, categoryID *int64, tagIDs []int64, page, pageSize int) ([]*domain.ArticleSummary, int64, error) {
	offset := (page - 1) * pageSize
	base := orm.NewSelector[domain.Article](r.db).OrderBy(orm.Desc("PublishedAt"))
	if categoryID != nil && *categoryID > 0 {
		base = base.Where(orm.C("CategoryID").Eq(*categoryID))
	}
	if len(tagIDs) > 0 {
		conds := make([]string, 0, len(tagIDs))
		args := make([]any, 0, len(tagIDs))
		for range tagIDs {
			conds = append(conds, "at.tag_id = ?")
		}
		for _, id := range tagIDs {
			args = append(args, id)
		}
		clause := "EXISTS (SELECT 1 FROM blog_article_tags at WHERE at.article_id = blog_article.id AND (" + strings.Join(conds, " OR ") + "))"
		base = base.Where(orm.Raw(clause, args...).AsPredicate())
	}
	rows, err := base.Limit(pageSize).Offset(offset).GetMulti(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListArticleSummariesFiltered 查询失败: %v", err)
		return nil, 0, err
	}
	summaries := make([]*domain.ArticleSummary, 0, len(rows))
	for _, a := range rows {
		if a == nil {
			continue
		}
		summaries = append(summaries, &domain.ArticleSummary{ID: a.ID, Title: a.Title})
	}
	// 统计总数
	cntSel := orm.NewSelector[aggregate.Result](r.db).From(orm.TableOf(&domain.Article{})).Select(orm.Count("ID").As("Count"))
	if categoryID != nil && *categoryID > 0 {
		cntSel = cntSel.Where(orm.C("CategoryID").Eq(*categoryID))
	}
	if len(tagIDs) > 0 {
		conds := make([]string, 0, len(tagIDs))
		args := make([]any, 0, len(tagIDs))
		for range tagIDs {
			conds = append(conds, "at.tag_id = ?")
		}
		for _, id := range tagIDs {
			args = append(args, id)
		}
		clause := "EXISTS (SELECT 1 FROM blog_article_tags at WHERE at.article_id = blog_article.id AND (" + strings.Join(conds, " OR ") + "))"
		cntSel = cntSel.Where(orm.Raw(clause, args...).AsPredicate())
	}
	cnt, err := cntSel.Get(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListArticleSummariesFiltered 统计失败: %v", err)
		return nil, 0, err
	}
	return summaries, cnt.Count, nil
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

func (r *ContentRepository) ListAllTags(ctx context.Context) ([]*domain.Tag, error) {
	list, err := orm.NewSelector[domain.Tag](r.db).OrderBy(orm.Asc("ID")).GetMulti(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: ListAllTags 查询失败: %v", err)
		return nil, err
	}
	return list, nil
}

func (r *ContentRepository) CountArticlesByTag(ctx context.Context, tagID int64) (int64, error) {
	cnt, err := orm.NewSelector[aggregate.Result](r.db).
		From(orm.TableOf(&domain.ArticleTag{})).
		Select(orm.Count("ID").As("count")).
		Where(orm.Raw("tag_id = ?", tagID).AsPredicate()).
		Get(ctx)
	if err != nil {
		logger.Log().Error("infrastructure: CountArticlesByTag 统计失败: %v", err)
		return 0, err
	}
	return cnt.Count, nil
}
