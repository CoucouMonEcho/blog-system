package clients

import (
	"context"
	"errors"
	"time"

	conf "blog-system/common/pkg/config"
	"blog-system/services/admin/application"
	"blog-system/services/admin/domain"
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

func NewContentClient(cfg *conf.AppConfig) *ContentClient {
	var reg registry.Registry
	if len(cfg.Registry.Endpoints) > 0 {
		if cli, err := clientv3.New(clientv3.Config{Endpoints: cfg.Registry.Endpoints, DialTimeout: 3 * time.Second}); err == nil {
			if r, err2 := regEtcd.NewRegistry(cli); err2 == nil {
				reg = r
			}
		}
	}
	c, _ := micro.NewClient(micro.ClientWithInsecure(), micro.ClientWithRegistry(reg, 3*time.Second))
	cc, _ := c.Dial(context.Background(), "content-grpc")
	return &ContentClient{cc: cc, cli: cpb.NewContentAdminServiceClient(cc)}
}

func (c *ContentClient) CreateArticle(ctx context.Context, a *domain.Article) error {
	_, err := c.cli.CreateArticle(ctx, &cpb.Article{Title: a.Title, Slug: a.Slug, Content: a.Content, Summary: a.Summary, Cover: a.Cover, AuthorId: a.AuthorID, CategoryId: a.CategoryID, Status: int32(a.Status), IsTop: a.IsTop, IsRecommend: a.IsRecommend})
	return err
}
func (c *ContentClient) UpdateArticle(ctx context.Context, a *domain.Article) error {
	_, err := c.cli.UpdateArticle(ctx, &cpb.Article{Id: a.ID, Title: a.Title, Slug: a.Slug, Content: a.Content, Summary: a.Summary, Cover: a.Cover, CategoryId: a.CategoryID, Status: int32(a.Status), IsTop: a.IsTop, IsRecommend: a.IsRecommend})
	return err
}
func (c *ContentClient) DeleteArticle(ctx context.Context, id int64) error {
	_, err := c.cli.DeleteArticle(ctx, &cpb.Id{Id: id})
	return err
}
func (c *ContentClient) ListArticles(ctx context.Context) ([]*domain.Article, int64, error) {
	resp, err := c.cli.ListArticles(ctx, &cpb.Empty{})
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
func (c *ContentClient) ListCategories(ctx context.Context) ([]*domain.Category, int64, error) {
	resp, err := c.cli.ListCategories(ctx, &cpb.Empty{})
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

// 标签管理
func (c *ContentClient) CreateTag(ctx context.Context, t *domain.Tag) error {
	_, err := c.cli.CreateTag(ctx, &cpb.Tag{Name: t.Name, Slug: t.Slug, Color: t.Color})
	return err
}
func (c *ContentClient) UpdateTag(ctx context.Context, t *domain.Tag) error {
	_, err := c.cli.UpdateTag(ctx, &cpb.Tag{Id: t.ID, Name: t.Name, Slug: t.Slug, Color: t.Color})
	return err
}
func (c *ContentClient) DeleteTag(ctx context.Context, id int64) error {
	_, err := c.cli.DeleteTag(ctx, &cpb.Id{Id: id})
	return err
}
func (c *ContentClient) ListTags(ctx context.Context) ([]*domain.Tag, error) {
	resp, err := c.cli.ListTags(ctx, &cpb.Empty{})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Tag, 0, len(resp.Tags))
	for _, t := range resp.Tags {
		out = append(out, &domain.Tag{ID: t.Id, Name: t.Name, Slug: t.Slug, Color: t.Color})
	}
	return out, nil
}
func (c *ContentClient) CountTags(ctx context.Context) (int64, error) {
	resp, err := c.cli.CountTags(ctx, &cpb.Empty{})
	if err != nil {
		return 0, err
	}
	return resp.Value, nil
}

var _ application.ContentClient = (*ContentClient)(nil)
