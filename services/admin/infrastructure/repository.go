package infrastructure

import (
	"blog-system/services/admin/domain"
	"context"

	"github.com/CoucouMonEcho/go-framework/orm"
)

type UserRepository struct{ db *orm.DB }
type ArticleRepository struct{ db *orm.DB }
type CategoryRepository struct{ db *orm.DB }

func NewUserRepository(db *orm.DB) *UserRepository         { return &UserRepository{db: db} }
func NewArticleRepository(db *orm.DB) *ArticleRepository   { return &ArticleRepository{db: db} }
func NewCategoryRepository(db *orm.DB) *CategoryRepository { return &CategoryRepository{db: db} }

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	return orm.NewInserter[domain.User](r.db).Values(u).Exec(ctx).Err()
}
func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	return orm.NewUpdater[domain.User](r.db).Set(orm.C("Username"), u.Username).Set(orm.C("Email"), u.Email).Set(orm.C("Password"), u.Password).Set(orm.C("Role"), u.Role).Set(orm.C("Avatar"), u.Avatar).Set(orm.C("Status"), u.Status).Set(orm.C("UpdatedAt"), u.UpdatedAt).Where(orm.C("Id").Eq(u.ID)).Exec(ctx).Err()
}
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return orm.NewUpdater[domain.User](r.db).Set(orm.C("Status"), 1).Where(orm.C("Id").Eq(id)).Exec(ctx).Err()
}
func (r *UserRepository) List(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	offset := (page - 1) * pageSize
	list, err := orm.NewSelector[domain.User](r.db).OrderBy(orm.Desc("CreatedAt")).Limit(pageSize).Offset(offset).GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	all, err := orm.NewSelector[domain.User](r.db).GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, int64(len(all)), nil
}

func (r *ArticleRepository) Create(ctx context.Context, a *domain.Article) error {
	return orm.NewInserter[domain.Article](r.db).Values(a).Exec(ctx).Err()
}
func (r *ArticleRepository) Update(ctx context.Context, a *domain.Article) error {
	return orm.NewUpdater[domain.Article](r.db).Set(orm.C("Title"), a.Title).Set(orm.C("Slug"), a.Slug).Set(orm.C("Content"), a.Content).Set(orm.C("Summary"), a.Summary).Set(orm.C("CategoryID"), a.CategoryID).Set(orm.C("Status"), a.Status).Set(orm.C("IsTop"), a.IsTop).Set(orm.C("IsRecommend"), a.IsRecommend).Set(orm.C("PublishedAt"), a.PublishedAt).Set(orm.C("UpdatedAt"), a.UpdatedAt).Where(orm.C("Id").Eq(a.ID)).Exec(ctx).Err()
}
func (r *ArticleRepository) Delete(ctx context.Context, id int64) error {
	return orm.NewUpdater[domain.Article](r.db).Set(orm.C("Status"), 2).Where(orm.C("Id").Eq(id)).Exec(ctx).Err()
}
func (r *ArticleRepository) List(ctx context.Context, page, pageSize int) ([]*domain.Article, int64, error) {
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

func (r *CategoryRepository) Create(ctx context.Context, c *domain.Category) error {
	return orm.NewInserter[domain.Category](r.db).Values(c).Exec(ctx).Err()
}
func (r *CategoryRepository) Update(ctx context.Context, c *domain.Category) error {
	return orm.NewUpdater[domain.Category](r.db).Set(orm.C("Name"), c.Name).Set(orm.C("Slug"), c.Slug).Set(orm.C("ParentID"), c.ParentID).Set(orm.C("Sort"), c.Sort).Where(orm.C("Id").Eq(c.ID)).Exec(ctx).Err()
}
func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	return orm.NewUpdater[domain.Category](r.db).Where(orm.C("Id").Eq(id)).Exec(ctx).Err()
}
func (r *CategoryRepository) List(ctx context.Context, page, pageSize int) ([]*domain.Category, int64, error) {
	offset := (page - 1) * pageSize
	list, err := orm.NewSelector[domain.Category](r.db).OrderBy(orm.Asc("Sort")).Limit(pageSize).Offset(offset).GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	all, err := orm.NewSelector[domain.Category](r.db).GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, int64(len(all)), nil
}
