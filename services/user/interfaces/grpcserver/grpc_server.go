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

func (s *GRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	u, token, err := s.app.Login(ctx, req.Username, req.Password)
	if err != nil {
		return &pb.LoginResponse{Code: 1, Message: err.Error()}, nil
	}
	avatar := ""
	if u.Avatar != nil && u.Avatar.Valid {
		avatar = u.Avatar.String
	}
	return &pb.LoginResponse{Code: 0, Message: "success", Token: token, User: &pb.User{
		Id:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Role:     u.Role,
		Avatar:   avatar,
		Status:   int32(u.Status),
	}}, nil
}

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

func (s *GRPCServer) UpdateUserStatus(ctx context.Context, req *pb.UpdateUserStatusRequest) (*pb.UpdateUserStatusResponse, error) {
	if err := s.app.ChangeUserStatus(ctx, req.UserId, int(req.Status)); err != nil {
		return &pb.UpdateUserStatusResponse{Code: 1, Message: err.Error()}, nil
	}
	return &pb.UpdateUserStatusResponse{Code: 0, Message: "success"}, nil
}

// Register 新建用户
func (s *GRPCServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	u, err := s.app.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return &pb.RegisterResponse{Code: 1, Message: err.Error()}, nil
	}
	return &pb.RegisterResponse{Code: 0, Message: "success", Data: &pb.User{Id: u.ID, Username: u.Username, Email: u.Email, Role: u.Role}}, nil
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
