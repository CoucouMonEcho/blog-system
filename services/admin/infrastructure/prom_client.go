package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"blog-system/services/admin/application"
)

// PrometheusClient 通过 Prometheus HTTP API 查询
type PrometheusClient struct {
	base string // http://prometheus:9090
	cli  *http.Client
}

func NewPrometheusClient(base string, timeout time.Duration) *PrometheusClient {
	if base == "" {
		base = "http://127.0.0.1:9090"
	}
	if timeout == 0 {
		timeout = 3 * time.Second
	}
	return &PrometheusClient{base: base, cli: &http.Client{Timeout: timeout}}
}

// doQueryRange 简化查询
func (p *PrometheusClient) doQueryRange(ctx context.Context, query, start, end, step string) (any, error) {
	u, _ := url.Parse(p.base + "/api/v1/query_range")
	q := u.Query()
	q.Set("query", query)
	q.Set("start", start)
	q.Set("end", end)
	q.Set("step", step)
	u.RawQuery = q.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := p.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var out struct {
		Status string `json:"status"`
		Data   any    `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Status != "success" {
		return nil, errors.New("prom query failed")
	}
	return out.Data, nil
}

func (p *PrometheusClient) ErrorRate(ctx context.Context, from, to, service string) (float64, error) {
	// 占位：实际可用 PromQL 如 sum(rate(http_requests_total{service="X",status=~"5.."}[5m])) / sum(rate(http_requests_total{service="X"}[5m]))
	return 0, nil
}

func (p *PrometheusClient) LatencyPercentile(ctx context.Context, from, to, service string) (map[string]float64, error) {
	// 占位：percentile over histogram metrics
	return map[string]float64{"p50": 0, "p90": 0, "p95": 0, "p99": 0}, nil
}

func (p *PrometheusClient) TopEndpoints(ctx context.Context, from, to, service string, topN int) ([]map[string]any, error) {
	// 占位：按 QPS 排序的 topN 接口
	return []map[string]any{}, nil
}

func (p *PrometheusClient) ActiveUsers(ctx context.Context, from, to string) (int64, error) {
	// 占位：如根据登录/访问行为去重计算活跃用户
	return 0, nil
}

var _ application.PromClient = (*PrometheusClient)(nil)
