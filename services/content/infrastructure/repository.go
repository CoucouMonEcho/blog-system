package infrastructure

import (
	"context"

	"blog-system/services/content/domain"

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
	list, err := orm.NewSelector[domain.Article](r.db).OrderBy(orm.Desc("PublishedAt")).Limit(pageSize).Offset(offset).GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	all, err := orm.NewSelector[domain.Article](r.db).GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, int64(len(all)), nil
}
