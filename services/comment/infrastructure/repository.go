package infrastructure

import (
	"blog-system/services/comment/domain"
	"context"
	"github.com/CoucouMonEcho/go-framework/orm"
)

type CommentRepository struct{ db *orm.DB }

func NewCommentRepository(db *orm.DB) *CommentRepository { return &CommentRepository{db: db} }

func (r *CommentRepository) Create(ctx context.Context, c *domain.Comment) error {
	res := orm.NewInserter[domain.Comment](r.db).
		Values(c).Exec(ctx)
	return res.Err()
}

func (r *CommentRepository) GetByID(ctx context.Context, id int64) (*domain.Comment, error) {
	return orm.NewSelector[domain.Comment](r.db).Where(orm.C("Id").Eq(id)).Get(ctx)
}
