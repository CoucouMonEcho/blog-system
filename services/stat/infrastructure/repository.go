package infrastructure

import (
	"blog-system/services/stat/domain"
	"context"
	"database/sql"

	"github.com/CoucouMonEcho/go-framework/orm"
)

type StatRepository struct{ db *orm.DB }

func NewStatRepository(db *orm.DB) *StatRepository { return &StatRepository{db: db} }

func (r *StatRepository) Incr(ctx context.Context, typ string, targetID int64, targetType string, userID *int64) error {
	m := &domain.Metric{Type: typ, TargetID: targetID, TargetType: targetType, Count: 1}
	if userID != nil {
		m.UserID = userID
	}
	// 使用 go-framework ORM 的 Upsert
	return orm.NewInserter[domain.Metric](r.db).
		Values(m).
		OnDuplicateKey().
		ConflictColumns("Type", "TargetID", "TargetType", "UserID").
		Update(
			// Count = Count + 1
			orm.Assign("Count", orm.Raw("Count + 1").AsPredicate()),
		).
		Exec(ctx).Err()
}

func (r *StatRepository) Get(ctx context.Context, typ string, targetID int64, targetType string, userID *int64) (int64, error) {
	q := orm.NewSelector[domain.Metric](r.db).
		Where(orm.C("Type").Eq(typ)).
		Where(orm.C("TargetID").Eq(targetID)).
		Where(orm.C("TargetType").Eq(targetType))
	if userID != nil {
		q = q.Where(orm.C("UserID").Eq(*userID))
	}
	m, err := q.Get(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return m.Count, nil
}
