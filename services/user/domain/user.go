package domain

import (
	"context"
	"database/sql"
	"time"
)

// User 用户领域模型
type User struct {
	ID        int64           `json:"id"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	Password  string          `json:"-"` // 不序列化密码
	Role      string          `json:"role"`
	Avatar    *sql.NullString `json:"avatar"`
	Status    int             `json:"status"` // 0:正常, 1:禁用
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// TableName 显式指定表名，避免命名策略差异导致的表名不匹配
func (User) TableName() string { return "blog_user" }

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, pageSize int) ([]*User, int64, error)
	UpdateStatus(ctx context.Context, id int64, status int) error
}

// UserService 用户领域服务
type UserService interface {
	Register(ctx context.Context, username, email, password string) (*User, error)
	Login(ctx context.Context, username, password string) (*User, error)
	GetUserInfo(ctx context.Context, id int64) (*User, error)
	UpdateUserInfo(ctx context.Context, id int64, updates map[string]interface{}) error
	ChangePassword(ctx context.Context, id int64, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, email string) error
}
