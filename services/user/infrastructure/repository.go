package infrastructure

import (
	"context"
	"database/sql"
	"errors"
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
	//TODO 规范化：去除前后空格，并按 MySQL collation 进行精确匹配
	uname := strings.TrimSpace(username)

	// 添加详细的调试日志
	logger.L().LogWithContext("user-service", "repository", "DEBUG", "sess is: %v", r.db)
	logger.L().LogWithContext("user-service", "repository", "DEBUG", "show tables is: %v",
		orm.RawQuery[domain.User](r.db, "show tables;").Exec(ctx))

	all, err := orm.RawQuery[domain.User](r.db, "select * from blog_user;").GetMulti(ctx)
	logger.L().LogWithContext("user-service", "repository", "DEBUG", "select all is: %v",
		all)

	logger.L().LogWithContext("user-service", "repository", "DEBUG", "Input username: %s", username)
	logger.L().LogWithContext("user-service", "repository", "DEBUG", "Trimmed username: %s", uname)
	logger.L().LogWithContext("user-service", "repository", "DEBUG", "Username type: %T, length: %d", uname, len(uname))

	// 构建查询
	selector := orm.NewSelector[domain.User](r.db).Where(orm.C("Username").Eq(uname))
	query, err := selector.Build()
	if err != nil {
		logger.L().LogWithContext("user-service", "repository", "ERROR", "Build error: %v", err)
		return nil, err
	}

	// 详细的查询信息
	logger.L().LogWithContext("user-service", "repository", "DEBUG", "Built query: %s", query.SQL)
	logger.L().LogWithContext("user-service", "repository", "DEBUG", "Query args count: %d", len(query.Args))
	for i, arg := range query.Args {
		logger.L().LogWithContext("user-service", "repository", "DEBUG", "Arg[%d]: %v (type: %T)", i, arg, arg)
	}

	// 执行查询
	user, err := selector.Get(ctx)
	if err != nil {
		logger.L().LogWithContext("user-service", "repository", "ERROR", "Query execution error: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	logger.L().LogWithContext("user-service", "repository", "DEBUG", "Query successful, user found: %v", user != nil)
	if user != nil {
		logger.L().LogWithContext("user-service", "repository", "DEBUG", "User ID: %d, Username: %s", user.ID, user.Username)
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
