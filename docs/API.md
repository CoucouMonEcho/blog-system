# API 文档（Gateway / User / Content / Admin / Stat）

本文档基于各服务的 `interfaces/httpserver/http.go` 最新实现，汇总网关可访问的 HTTP API。网关会按服务前缀将请求代理到后端服务。

通用返回结构：

```json
{ "code": 0, "message": "success", "data": {} }
```

## 网关（gateway）
- 健康检查: `GET /health`
- 代理入口: `GET|POST /api/*`
- 鉴权：除 `GET /health` 与 `POST /api/user/login` 外，其他 `/api/*` 路径均要求 `Authorization: Bearer <token>`。

### 路由前缀与后端服务映射
- 用户：`/api/user/**` → user-service `/api/**`
- 内容：`/api/content/**` → content-service `/api/**`
- 管理：`/api/admin/**` → admin-service `/api/**`
- 统计：`/api/stat/**` → stat-service `/api/**`

---

## 用户（user）
- 健康检查: `GET /api/user/health`

### 登录
- `POST /api/user/login`
- 请求头：`Content-Type: application/json`
- 请求体：
```json
{ "username": "alice", "password": "secret" }
```
- 响应体：
```json
{ "code": 0, "message": "success", "data": { "token": "<jwt>", "user": { "id": 1, "username": "alice" } } }
```

### 认证接口（需要 JWT）
- 公共请求头：
```
Authorization: Bearer <token>
Content-Type: application/json
```
- 获取用户信息：`GET /api/user/info/:user_id`
  - 响应：
  ```json
  { "code": 0, "message": "success", "data": { "id": 1, "username": "alice", "email": "a@b.com" } }
  ```
- 更新资料：`POST /api/user/update`
  - 请求体（任意字段可选）：
  ```json
  { "username": "alice2", "email": "a2@b.com", "avatar": "https://..." }
  ```
  - 响应：`{ "code": 0, "message": "success", "data": null }`
- 修改密码：`POST /api/user/password`
  - 请求体：
  ```json
  { "old_password": "old", "new_password": "new-123456" }
  ```
  - 响应：`{ "code": 0, "message": "success", "data": null }`

---

## 内容（content）
- 健康检查: `GET /api/content/health`
- 通用请求头：`Content-Type: application/json`（仅 POST）

### 文章详情
- `GET /api/content/article/:article_id`
- 响应：
```json
{ "code": 0, "message": "success", "data": { "id": 1, "title": "T", "content": "..." } }
```

### 文章摘要列表（支持过滤）
- `GET /api/content/article/list?page=&page_size=&category_id=&tag_ids=`
- 说明：
  - `category_id` 可选；`tag_ids` 逗号分隔；`page/page_size` 分页
- 响应示例（可选字段仅在非空时返回）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "title": "Intro to Go",
        "summary": "Go 是一门开源编程语言...",
        "author_id": 1001,
        "category": { "id": 2, "name": "后端", "slug": "backend" },
        "tags": [
          { "id": 1, "name": "Go", "slug": "go", "color": "#00ADD8" },
          { "id": 3, "name": "并发", "slug": "concurrency", "color": "#666" }
        ],
        "cover_url": "https://cdn.example.com/covers/go-intro.jpg"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

### 关键词搜索（返回摘要列表）
- `GET /api/content/article/search?q=&page=&page_size=`
- 响应结构与“文章摘要列表”相同。

### 分类列表（单层，全量）
- `GET /api/content/category/list`
- 响应：
```json
{ "code": 0, "message": "success", "data": [{"id":1,"name":"A","slug":"a","sort":10}] }
```

### 标签列表（全量，含文章计数）
- `GET /api/content/tag/list`
- 响应：
```json
{ "code": 0, "message": "success", "data": [
  { "id": 1, "name": "Go", "slug": "go", "color": "#00ADD8", "count": 42 },
  { "id": 2, "name": "Web", "slug": "web", "color": "#333333", "count": 18 }
] }
```

---

## 管理（admin）
- 健康检查: `GET /api/admin/health`
- 认证：所有 `/api/admin/**` 接口均需 `Authorization: Bearer <token>`（请先通过 `/api/user/login` 获取）。

### 用户管理
- 分页查询：`GET /api/admin/users?page=&page_size=`
  - 响应：`{ code,message,data:{ list:[], total:0, page:1, page_size:10 } }`
- 新增：`POST /api/admin/users`
  - 请求头：`Content-Type: application/json`
  - 请求体：
  ```json
  { "username":"u","email":"u@a.com","password":"p","role":"admin","avatar":"","status":1 }
  ```
- 修改：`POST /api/admin/users/update/:id`
  - 请求体：`{"email":"u2@a.com","role":"user"}`
- 删除：`POST /api/admin/users/delete/:id`

### 文章管理（全量列表）
- 列表（全量）：`GET /api/admin/articles`
  - 响应：`{ code,message,data:{ list:[], total:0, page:1, page_size:<len(list)> } }`
- 新增：`POST /api/admin/articles`
  - 请求头：`Content-Type: application/json`
  - 请求体：
  ```json
  { "title":"T","slug":"t","content":"...","summary":"...","author_id":1,"category_id":2,"status":0,"is_top":false,"is_recommend":false }
  ```
- 修改：`POST /api/admin/articles/update/:id`
  - 请求体：`{"title":"T2","category_id":3}`
- 删除：`POST /api/admin/articles/delete/:id`

### 分类管理（全量列表）
- 列表（全量）：`GET /api/admin/categories`
  - 响应：`{ code,message,data:{ list:[], total:0, page:1, page_size:<len(list)> } }`
- 新增：`POST /api/admin/categories`
  - 请求头：`Content-Type: application/json`
  - 请求体：`{"name":"后端","slug":"backend","sort":10}`
- 修改：`POST /api/admin/categories/update/:id`
  - 请求体：`{"name":"服务端","sort":20}`
- 删除：`POST /api/admin/categories/delete/:id`
- 分类（扁平树展示）：`GET /api/admin/categories/tree`
  - 响应：`{ code,message,data:{ list:[{"id":1,"name":"A","slug":"a","sort":10}], total:<n>, page:1, page_size:<len(list)> } }`

### 仪表盘（admin）
- 概览：`GET /api/admin/stat/overview`
  - 响应：
  ```json
  { "code":0, "message":"success", "data": {"pv_today":123, "uv_today":45, "online_users":12, "article_total": 200, "category_total": 20, "error_5xx_last_1h": 0} }
  ```
- PV 时间序列：`GET /api/admin/stat/pv_timeseries?from=2025-08-01T00:00:00Z&to=2025-08-01T12:00:00Z&interval=1h`
  - 响应：`{ code:0, data: [{"ts":1722470400, "value": 100}] }`
- 错误率：`GET /api/admin/stat/error_rate?from=...&to=...&service=admin`
  - 响应：`{ code:0, data: { "error_rate": 0.01 } }`
- 延迟分位：`GET /api/admin/stat/latency_percentile?from=...&to=...&service=user`
  - 响应：`{ code:0, data: { "p50":10, "p90":30, "p95":60, "p99":120 } }`
- Top 接口：`GET /api/admin/stat/top_endpoints?from=...&to=...&service=content[&top=10]`
  - 响应：`{ code:0, data: [{"path":"/api/article/list","qps": 12.3}] }`
- 活跃用户：`GET /api/admin/stat/active_users?from=...&to=...`
  - 响应：`{ code:0, data: { "active_users": 1234 } }`

---

## 统计（stat）
- 健康检查: `GET /api/stat/health`

- 自增统计：`POST /api/stat/incr?type=view&target_id=1&target_type=article[&user_id=2]`
  - 请求头：`Content-Type: application/json`
  - 请求体：空（参数在 query）
  - 响应：`{ "code": 0, "message": "success", "data": null }`

- 获取统计：`GET /api/stat/get?type=view&target_id=1&target_type=article[&user_id=2]`
  - 响应：`{ "code": 0, "message": "success", "data": { "value": 123 } }`

---

# 十条典型用例
1. 登录：`POST /api/user/login` → 返回 token
2. 获取用户信息：`GET /api/user/info/1` → 返回 user
3. 文章详情：`GET /api/content/article/1` → 返回 article
4. 文章摘要列表（过滤）：`GET /api/content/article/list?category_id=2&tag_ids=1,3&page=1&page_size=10`
5. 关键词搜索：`GET /api/content/article/search?q=go&page=1&page_size=10`
6. 分类列表（全量）：`GET /api/content/category/list`
7. 标签列表：`GET /api/content/tag/list`（每项含 `count`）
8. 管理端新增文章：`POST /api/admin/articles`
9. 管理端更新文章：`POST /api/admin/articles/update/1`
10. 增加统计值：`POST /api/stat/incr?type=like&target_id=1&target_type=article&user_id=2`
