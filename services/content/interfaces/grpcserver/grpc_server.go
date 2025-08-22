package api

import (
	"context"

	"blog-system/services/content/application"
	"blog-system/services/content/domain"
	pb "blog-system/services/content/proto"
)

type AdminGRPCServer struct {
	pb.UnimplementedContentAdminServiceServer
	app *application.ContentAppService
}

func NewAdminGRPCServer(app *application.ContentAppService) *AdminGRPCServer {
	return &AdminGRPCServer{app: app}
}

// Article
func (s *AdminGRPCServer) CreateArticle(ctx context.Context, req *pb.Article) (*pb.Empty, error) {
	a := &domain.Article{Title: req.Title, Slug: req.Slug, Content: req.Content, Summary: req.Summary, AuthorID: req.AuthorId, CategoryID: req.CategoryId, Status: int(req.Status), IsTop: req.IsTop, IsRecommend: req.IsRecommend}
	_, err := s.app.Create(ctx, a)
	return &pb.Empty{}, err
}
func (s *AdminGRPCServer) UpdateArticle(ctx context.Context, req *pb.Article) (*pb.Empty, error) {
	a := &domain.Article{ID: req.Id, Title: req.Title, Slug: req.Slug, Content: req.Content, Summary: req.Summary, CategoryID: req.CategoryId, Status: int(req.Status), IsTop: req.IsTop, IsRecommend: req.IsRecommend}
	return &pb.Empty{}, s.app.Update(ctx, a)
}
func (s *AdminGRPCServer) DeleteArticle(ctx context.Context, req *pb.Id) (*pb.Empty, error) {
	return &pb.Empty{}, s.app.Delete(ctx, req.Id)
}
func (s *AdminGRPCServer) ListArticles(ctx context.Context, req *pb.Page) (*pb.ArticleListResponse, error) {
	list, total, err := s.app.ListSummaries(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}
	// 仅返回 ID 与 Title，其他字段可按需补充
	out := make([]*pb.Article, 0, len(list))
	for _, it := range list {
		out = append(out, &pb.Article{Id: it.ID, Title: it.Title})
	}
	return &pb.ArticleListResponse{Data: out, Total: total}, nil
}
func (s *AdminGRPCServer) CountArticles(ctx context.Context, _ *pb.Empty) (*pb.Count, error) {
	// 简化：调用列表+total（若需精确 COUNT，可在 repo 增加 Count）
	_, total, err := s.app.ListSummaries(ctx, 1, 1)
	if err != nil {
		return nil, err
	}
	return &pb.Count{Value: total}, nil
}

// Category
func (s *AdminGRPCServer) CreateCategory(ctx context.Context, req *pb.Category) (*pb.Empty, error) {
	c := &domain.Category{Name: req.Name, Slug: req.Slug, Sort: int(req.Sort)}
	return &pb.Empty{}, s.app.UpdateCategory(ctx, c) // Create 复用 UpdateCategory 的缓存失效逻辑
}
func (s *AdminGRPCServer) UpdateCategory(ctx context.Context, req *pb.Category) (*pb.Empty, error) {
	c := &domain.Category{ID: req.Id, Name: req.Name, Slug: req.Slug, Sort: int(req.Sort)}
	return &pb.Empty{}, s.app.UpdateCategory(ctx, c)
}
func (s *AdminGRPCServer) DeleteCategory(ctx context.Context, req *pb.Id) (*pb.Empty, error) {
	return &pb.Empty{}, s.app.DeleteCategory(ctx, req.Id)
}
func (s *AdminGRPCServer) ListCategories(ctx context.Context, req *pb.Page) (*pb.CategoryListResponse, error) {
	list, total, err := s.app.ListCategories(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}
	out := make([]*pb.Category, 0, len(list))
	for _, c := range list {
		out = append(out, &pb.Category{Id: c.ID, Name: c.Name, Slug: c.Slug, Sort: int32(c.Sort)})
	}
	return &pb.CategoryListResponse{Data: out, Total: total}, nil
}
func (s *AdminGRPCServer) CountCategories(ctx context.Context, _ *pb.Empty) (*pb.Count, error) {
	// 简化：page=1 size=1 获取 total
	_, total, err := s.app.ListCategories(ctx, 1, 1)
	if err != nil {
		return nil, err
	}
	return &pb.Count{Value: total}, nil
}
