package grpcserver

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
	a := &domain.Article{Title: req.Title, Slug: req.Slug, Content: req.Content, Summary: req.Summary, Cover: req.Cover, AuthorID: req.AuthorId, CategoryID: req.CategoryId, Status: int(req.Status), IsTop: req.IsTop, IsRecommend: req.IsRecommend}
	_, err := s.app.Create(ctx, a)
	return &pb.Empty{}, err
}
func (s *AdminGRPCServer) UpdateArticle(ctx context.Context, req *pb.Article) (*pb.Empty, error) {
	a := &domain.Article{ID: req.Id, Title: req.Title, Slug: req.Slug, Content: req.Content, Summary: req.Summary, Cover: req.Cover, CategoryID: req.CategoryId, Status: int(req.Status), IsTop: req.IsTop, IsRecommend: req.IsRecommend}
	return &pb.Empty{}, s.app.Update(ctx, a)
}
func (s *AdminGRPCServer) DeleteArticle(ctx context.Context, req *pb.Id) (*pb.Empty, error) {
	return &pb.Empty{}, s.app.Delete(ctx, req.Id)
}
func (s *AdminGRPCServer) ListArticles(ctx context.Context, _ *pb.Empty) (*pb.ArticleListResponse, error) {
	list, total, err := s.app.ListAllArticles(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*pb.Article, 0, len(list))
	for _, it := range list {
		out = append(out, &pb.Article{Id: it.ID, Title: it.Title})
	}
	return &pb.ArticleListResponse{Data: out, Total: total}, nil
}
func (s *AdminGRPCServer) CountArticles(ctx context.Context, _ *pb.Empty) (*pb.Count, error) {
	val, err := s.app.CountArticles(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.Count{Value: val}, nil
}

// Category（全量）
func (s *AdminGRPCServer) CreateCategory(ctx context.Context, req *pb.Category) (*pb.Empty, error) {
	c := &domain.Category{Name: req.Name, Slug: req.Slug, Sort: int(req.Sort)}
	return &pb.Empty{}, s.app.UpdateCategory(ctx, c)
}
func (s *AdminGRPCServer) UpdateCategory(ctx context.Context, req *pb.Category) (*pb.Empty, error) {
	c := &domain.Category{ID: req.Id, Name: req.Name, Slug: req.Slug, Sort: int(req.Sort)}
	return &pb.Empty{}, s.app.UpdateCategory(ctx, c)
}
func (s *AdminGRPCServer) DeleteCategory(ctx context.Context, req *pb.Id) (*pb.Empty, error) {
	return &pb.Empty{}, s.app.DeleteCategory(ctx, req.Id)
}
func (s *AdminGRPCServer) ListCategories(ctx context.Context, _ *pb.Empty) (*pb.CategoryListResponse, error) {
	list, err := s.app.ListAllCategories(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*pb.Category, 0, len(list))
	for _, c := range list {
		out = append(out, &pb.Category{Id: c.ID, Name: c.Name, Slug: c.Slug, Sort: int32(c.Sort)})
	}
	return &pb.CategoryListResponse{Data: out, Total: int64(len(out))}, nil
}
func (s *AdminGRPCServer) CountCategories(ctx context.Context, _ *pb.Empty) (*pb.Count, error) {
	val, err := s.app.CountCategories(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.Count{Value: val}, nil
}

// Tag （全量）
func (s *AdminGRPCServer) CreateTag(ctx context.Context, req *pb.Tag) (*pb.Empty, error) {
	t := &domain.Tag{Name: req.Name, Slug: req.Slug, Color: req.Color}
	return &pb.Empty{}, s.app.CreateTag(ctx, t)
}
func (s *AdminGRPCServer) UpdateTag(ctx context.Context, req *pb.Tag) (*pb.Empty, error) {
	t := &domain.Tag{ID: req.Id, Name: req.Name, Slug: req.Slug, Color: req.Color}
	return &pb.Empty{}, s.app.UpdateTag(ctx, t)
}
func (s *AdminGRPCServer) DeleteTag(ctx context.Context, req *pb.Id) (*pb.Empty, error) {
	return &pb.Empty{}, s.app.DeleteTag(ctx, req.Id)
}
func (s *AdminGRPCServer) ListTags(ctx context.Context, _ *pb.Empty) (*pb.TagListResponse, error) {
	tags, _, err := s.app.ListAllTags(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*pb.Tag, 0, len(tags))
	for _, t := range tags {
		out = append(out, &pb.Tag{Id: t.ID, Name: t.Name, Slug: t.Slug, Color: t.Color})
	}
	return &pb.TagListResponse{Tags: out}, nil
}
func (s *AdminGRPCServer) CountTags(ctx context.Context, _ *pb.Empty) (*pb.Count, error) {
	val, err := s.app.CountTags(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.Count{Value: val}, nil
}
