package grpcserver

import (
	"context"

	"blog-system/services/user/application"
	pb "blog-system/services/user/proto"
)

type GRPCServer struct {
	pb.UnimplementedUserServiceServer
	app *application.UserAppService
}

func NewGRPCServer(app *application.UserAppService) *GRPCServer { return &GRPCServer{app: app} }

// Register 新建用户
func (s *GRPCServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	u, err := s.app.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return &pb.RegisterResponse{Code: 1, Message: err.Error()}, nil
	}
	return &pb.RegisterResponse{Code: 0, Message: "success", Data: &pb.User{Id: u.ID, Username: u.Username, Email: u.Email, Role: u.Role}}, nil
}

// ListUsers 用户列表
func (s *GRPCServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	list, total, err := s.app.ListUsers(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return &pb.ListUsersResponse{Code: 1, Message: err.Error()}, nil
	}
	out := make([]*pb.User, 0, len(list))
	for _, u := range list {
		avatar := ""
		if u.Avatar != nil && u.Avatar.Valid {
			avatar = u.Avatar.String
		}
		out = append(out, &pb.User{Id: u.ID, Username: u.Username, Email: u.Email, Role: u.Role, Avatar: avatar, Status: int32(u.Status)})
	}
	return &pb.ListUsersResponse{Code: 0, Message: "success", Data: out, Total: total}, nil
}

// UpdateUserStatus 更新用户状态
func (s *GRPCServer) UpdateUserStatus(ctx context.Context, req *pb.UpdateUserStatusRequest) (*pb.UpdateUserStatusResponse, error) {
	if err := s.app.ChangeUserStatus(ctx, req.UserId, int(req.Status)); err != nil {
		return &pb.UpdateUserStatusResponse{Code: 1, Message: err.Error()}, nil
	}
	return &pb.UpdateUserStatusResponse{Code: 0, Message: "success"}, nil
}

// UpdateUserInfo 更新用户信息
func (s *GRPCServer) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UpdateUserInfoResponse, error) {
	updates := map[string]any{}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if err := s.app.UpdateUserInfo(ctx, req.UserId, updates); err != nil {
		return &pb.UpdateUserInfoResponse{Code: 1, Message: err.Error()}, nil
	}
	return &pb.UpdateUserInfoResponse{Code: 0, Message: "success"}, nil
}

// GetUserInfo 获取用户信息
func (s *GRPCServer) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.GetUserInfoResponse, error) {
	u, err := s.app.GetUserInfo(ctx, req.UserId)
	if err != nil {
		return &pb.GetUserInfoResponse{Code: 1, Message: err.Error()}, nil
	}
	avatar := ""
	if u.Avatar != nil && u.Avatar.Valid {
		avatar = u.Avatar.String
	}
	return &pb.GetUserInfoResponse{Code: 0, Message: "success", Data: &pb.User{Id: u.ID, Username: u.Username, Email: u.Email, Role: u.Role, Avatar: avatar, Status: int32(u.Status)}}, nil
}

// ChangePassword 修改密码
func (s *GRPCServer) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	if err := s.app.ChangePassword(ctx, req.UserId, req.OldPassword, req.NewPassword); err != nil {
		return &pb.ChangePasswordResponse{Code: 1, Message: err.Error()}, nil
	}
	return &pb.ChangePasswordResponse{Code: 0, Message: "success"}, nil
}

// ResetPassword 重置密码
func (s *GRPCServer) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	if err := s.app.ResetPassword(ctx, req.Email); err != nil {
		return &pb.ResetPasswordResponse{Code: 1, Message: err.Error()}, nil
	}
	return &pb.ResetPasswordResponse{Code: 0, Message: "success"}, nil
}
