package infrastructure

import (
	"context"
	"strings"

	"blog-system/common/pkg/logger"
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
	res := orm.NewInserter[domain.User](r.db).
		Values(user).
		Exec(ctx)
	if err := res.Err(); err != nil {
		logger.Log().Error("repository: Create 用户失败: user=%+v err=%v", user, err)
		return err
	}
	return nil
}

// FindByID 根据ID查找用户
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := orm.NewSelector[domain.User](r.db).Where(orm.C("ID").Eq(id)).Get(ctx)
	if err != nil {
		logger.Log().Warn("repository: FindByID 查询失败: id=%d err=%v", id, err)
		return nil, err
	}
	return user, nil
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := orm.NewSelector[domain.User](r.db).Where(orm.C("Username").Eq(strings.TrimSpace(username))).Get(ctx)
	if err != nil {
		logger.Log().Warn("repository: FindByUsername 查询失败: username=%s err=%v", username, err)
		return nil, err
	}
	return user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := orm.NewSelector[domain.User](r.db).Where(orm.C("Email").Eq(email)).Get(ctx)
	if err != nil {
		//if errors.Is(err, sql.ErrNoRows) {
		//	return nil, errors.New("用户不存在")
		//}
		logger.Log().Warn("repository: FindByEmail 查询失败: email=%s err=%v", email, err)
		return nil, err
	}
	return user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	if err := orm.NewUpdater[domain.User](r.db).
		Set(orm.C("Username"), user.Username).
		Set(orm.C("Email"), user.Email).
		Set(orm.C("Password"), user.Password).
		Set(orm.C("Role"), user.Role).
		Set(orm.C("Avatar"), user.Avatar).
		Set(orm.C("Status"), user.Status).
		Set(orm.C("UpdatedAt"), user.UpdatedAt).
		Where(orm.C("ID").Eq(user.ID)).
		Exec(ctx).Err(); err != nil {
		logger.Log().Error("repository: Update 用户失败: id=%d err=%v", user.ID, err)
		return err
	}
	return nil
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	if err := orm.NewUpdater[domain.User](r.db).Set(orm.C("Status"), 1).Where(orm.C("ID").Eq(id)).Exec(ctx).Err(); err != nil {
		logger.Log().Error("repository: Delete 用户失败: id=%d err=%v", id, err)
		return err
	}
	return nil
}

// List 用户列表
func (r *UserRepository) List(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	offset := (page - 1) * pageSize
	users, err := orm.NewSelector[domain.User](r.db).OrderBy(orm.Desc("CreatedAt")).Limit(pageSize).Offset(offset).GetMulti(ctx)
	if err != nil {
		logger.Log().Warn("repository: List 查询列表失败: page=%d size=%d err=%v", page, pageSize, err)
		return nil, 0, err
	}

	total, err := orm.NewSelector[int64](r.db).From(orm.TableOf(&domain.User{})).Select(orm.Count("1")).Get(ctx)
	if err != nil {
		logger.Log().Warn("repository: List 统计总数失败: err=%v", err)
		return nil, 0, err
	}

	return users, *total, nil
}

// UpdateStatus 更新用户状态
func (r *UserRepository) UpdateStatus(ctx context.Context, id int64, status int) error {
	if err := orm.NewUpdater[domain.User](r.db).Set(orm.C("Status"), status).Where(orm.C("ID").Eq(id)).Exec(ctx).Err(); err != nil {
		logger.Log().Error("repository: UpdateStatus 失败: id=%d status=%d err=%v", id, status, err)
		return err
	}
	return nil
}
