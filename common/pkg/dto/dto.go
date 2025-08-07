package dto

// PageRequest 分页请求
type PageRequest struct {
	Page     int `json:"page" form:"page" binding:"required,min=1"`
	PageSize int `json:"page_size" form:"page_size" binding:"required,min=1,max=100"`
}

// PageResponse 分页响应
type PageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// Response 通用响应
type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

// Success 成功响应
func Success[T any](data T) Response[T] {
	return Response[T]{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// SuccessNil 空响应
func SuccessNil[T string]() Response[T] {
	return Response[T]{
		Code:    0,
		Message: "success",
	}
}

// Error 错误响应
func Error(code int, message string) Response[any] {
	return Response[any]{
		Code:    code,
		Message: message,
	}
}
