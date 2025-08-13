package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"blog-system/services/user/domain"

	"github.com/CoucouMonEcho/go-framework/orm"
)

var _ domain.UserRepository = &UserRepository{}

// UserRepository 用户仓储实现
type UserRepository struct {
	db *orm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *orm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	//TODO 显式指定列，避免 ORM 因零值字段被省略导致生成无列 INSERT
	res := orm.NewInserter[domain.User](r.db).
		Columns("Username", "Email", "Password", "Role", "Avatar", "Status", "CreatedAt", "UpdatedAt").
		Values(user).
		Exec(ctx)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

// FindByID 根据ID查找用户
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := orm.NewSelector[domain.User](r.db).Where(orm.C("Id").Eq(id)).Get(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return user, nil
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := orm.NewSelector[domain.User](r.db).Where(orm.C("Username").Eq(username)).Get(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := orm.NewSelector[domain.User](r.db).Where(orm.C("Email").Eq(email)).Get(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	return orm.NewUpdater[domain.User](r.db).Set(orm.C("Username"), user.Username).Set(orm.C("Email"), user.Email).Set(orm.C("Password"), user.Password).Set(orm.C("Role"), user.Role).Set(orm.C("Avatar"), user.Avatar).Set(orm.C("Status"), user.Status).Set(orm.C("UpdatedAt"), user.UpdatedAt).Where(orm.C("Id").Eq(user.ID)).Exec(ctx).Err()
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return orm.NewUpdater[domain.User](r.db).Set(orm.C("Status"), 1).Where(orm.C("Id").Eq(id)).Exec(ctx).Err()
}

// List 用户列表
func (r *UserRepository) List(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	offset := (page - 1) * pageSize
	users, err := orm.NewSelector[domain.User](r.db).OrderBy(orm.Desc("CreatedAt")).Limit(pageSize).Offset(offset).GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 获取总数 - 使用单独的查询
	countSelector := orm.NewSelector[domain.User](r.db)
	countUsers, err := countSelector.GetMulti(ctx)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(countUsers))

	return users, total, nil
}

// UpdateStatus 更新用户状态
func (r *UserRepository) UpdateStatus(ctx context.Context, id int64, status int) error {
	return orm.NewUpdater[domain.User](r.db).Set(orm.C("Status"), status).Where(orm.C("Id").Eq(id)).Exec(ctx).Err()
}
