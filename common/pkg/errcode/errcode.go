package errcode

// 系统级错误码
const (
	OK          = 0
	ErrInternal = 10001 + iota
	ErrParam
	ErrUnauthorized
	ErrNotFound
	ErrForbidden
	ErrConflict
)

// 管理服务错误码
const (
	ErrAdminForbidden = 20001 + iota
	ErrAdminNotFound
)

// 统计服务错误码
const (
	ErrStatNotFound = 30001 + iota
)

// 用户服务错误码
const (
	ErrUserNotFound = 40001 + iota
	ErrUserExists
	ErrPasswordInvalid
	ErrTokenInvalid
	ErrTokenExpired
)

// 内容服务错误码
const (
	ErrArticleNotFound = 50001 + iota
	ErrTagNotFound
	ErrCategoryNotFound
)

// ErrorMessage 错误码对应的消息
var ErrorMessage = map[int]string{
	OK:              "成功",
	ErrInternal:     "内部服务器错误",
	ErrParam:        "参数错误",
	ErrUnauthorized: "未授权",
	ErrNotFound:     "资源不存在",
	ErrForbidden:    "禁止访问",
	ErrConflict:     "资源冲突",

	ErrAdminForbidden: "管理员权限不足",
	ErrAdminNotFound:  "管理员不存在",

	ErrStatNotFound: "统计数据不存在",

	ErrUserNotFound:    "用户不存在",
	ErrUserExists:      "用户已存在",
	ErrPasswordInvalid: "密码错误",
	ErrTokenInvalid:    "令牌无效",
	ErrTokenExpired:    "令牌已过期",

	ErrArticleNotFound:  "文章不存在",
	ErrTagNotFound:      "标签不存在",
	ErrCategoryNotFound: "分类不存在",
}

// GetMessage 获取错误消息
func GetMessage(code int) string {
	if msg, ok := ErrorMessage[code]; ok {
		return msg
	}
	return "未知错误"
}
