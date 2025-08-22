package clients

import (
	"context"
	"errors"
	"time"

	"blog-system/services/admin/application"
	"blog-system/services/admin/domain"
	"blog-system/services/admin/infrastructure"
	cpb "blog-system/services/content/proto"

	micro "github.com/CoucouMonEcho/go-framework/micro"
	"github.com/CoucouMonEcho/go-framework/micro/registry"
	regEtcd "github.com/CoucouMonEcho/go-framework/micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type ContentClient struct {
	cc  *grpc.ClientConn
	cli cpb.ContentAdminServiceClient
}

func NewContentClient(cfg *infrastructure.AppConfig) *ContentClient {
	var reg registry.Registry
	if len(cfg.Registry.Endpoints) > 0 {
		if cli, err := clientv3.New(clientv3.Config{Endpoints: cfg.Registry.Endpoints, DialTimeout: 3 * time.Second}); err == nil {
			if r, err2 := regEtcd.NewRegistry(cli); err2 == nil {
				reg = r
			}
		}
	}
	c, _ := micro.NewClient(micro.ClientWithInsecure(), micro.ClientWithRegistry(reg, 3*time.Second))
	cc, _ := c.Dial(context.Background(), "content-service")
	return &ContentClient{cc: cc, cli: cpb.NewContentAdminServiceClient(cc)}
}

func (c *ContentClient) CreateArticle(ctx context.Context, a *domain.Article) error {
	_, err := c.cli.CreateArticle(ctx, &cpb.Article{Title: a.Title, Slug: a.Slug, Content: a.Content, Summary: a.Summary, AuthorId: a.AuthorID, CategoryId: a.CategoryID, Status: int32(a.Status), IsTop: a.IsTop, IsRecommend: a.IsRecommend})
	return err
}
func (c *ContentClient) UpdateArticle(ctx context.Context, a *domain.Article) error {
	_, err := c.cli.UpdateArticle(ctx, &cpb.Article{Id: a.ID, Title: a.Title, Slug: a.Slug, Content: a.Content, Summary: a.Summary, CategoryId: a.CategoryID, Status: int32(a.Status), IsTop: a.IsTop, IsRecommend: a.IsRecommend})
	return err
}
func (c *ContentClient) DeleteArticle(ctx context.Context, id int64) error {
	_, err := c.cli.DeleteArticle(ctx, &cpb.Id{Id: id})
	return err
}
func (c *ContentClient) ListArticles(ctx context.Context, page, pageSize int) ([]*domain.Article, int64, error) {
	resp, err := c.cli.ListArticles(ctx, &cpb.Page{Page: int32(page), PageSize: int32(pageSize)})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*domain.Article, 0, len(resp.Data))
	for _, a := range resp.Data {
		out = append(out, &domain.Article{ID: a.Id, Title: a.Title})
	}
	return out, resp.Total, nil
}
func (c *ContentClient) CountArticles(ctx context.Context) (int64, error) {
	resp, err := c.cli.CountArticles(ctx, &cpb.Empty{})
	if err != nil {
		return 0, err
	}
	return resp.Value, nil
}

func (c *ContentClient) CreateCategory(ctx context.Context, cat *domain.Category) error {
	_, err := c.cli.CreateCategory(ctx, &cpb.Category{Name: cat.Name, Slug: cat.Slug, Sort: int32(cat.Sort)})
	return err
}
func (c *ContentClient) UpdateCategory(ctx context.Context, cat *domain.Category) error {
	_, err := c.cli.UpdateCategory(ctx, &cpb.Category{Id: cat.ID, Name: cat.Name, Slug: cat.Slug, Sort: int32(cat.Sort)})
	return err
}
func (c *ContentClient) DeleteCategory(ctx context.Context, id int64) error {
	_, err := c.cli.DeleteCategory(ctx, &cpb.Id{Id: id})
	return err
}
func (c *ContentClient) ListCategories(ctx context.Context, page, pageSize int) ([]*domain.Category, int64, error) {
	resp, err := c.cli.ListCategories(ctx, &cpb.Page{Page: int32(page), PageSize: int32(pageSize)})
	if err != nil {
		return nil, 0, err
	}
	if resp == nil {
		return nil, 0, errors.New("empty response")
	}
	out := make([]*domain.Category, 0, len(resp.Data))
	for _, c0 := range resp.Data {
		out = append(out, &domain.Category{ID: c0.Id, Name: c0.Name, Slug: c0.Slug, Sort: int(c0.Sort)})
	}
	return out, resp.Total, nil
}
func (c *ContentClient) CountCategories(ctx context.Context) (int64, error) {
	resp, err := c.cli.CountCategories(ctx, &cpb.Empty{})
	if err != nil {
		return 0, err
	}
	return resp.Value, nil
}

var _ application.ContentClient = (*ContentClient)(nil)
