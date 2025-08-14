# API 文档（Gateway / User / Content / Admin / Stat）

本文档汇总各服务的 HTTP API，网关通过 `configs/gateway.yaml` 中的前缀进行路由转发（service://）。示例仅展示关键字段，具体响应见通用响应结构。

通用返回结构：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

## 网关（gateway）
- 健康检查: `GET /health`
- 代理入口: `GET|POST /api/*`

通用请求头（代理透传）：
```
Authorization: Bearer <token>   # 可选，透传到后端
Content-Type: application/json   # POST/PUT 建议
```

## 用户（user）
- 健康检查: `GET /health`

### 登录
- `POST /api/login`
- 请求头：
```
Content-Type: application/json
```
- 请求体：
```json
{ "username": "alice", "password": "secret" }
```
- 响应体：
```json
{ "code": 0, "message": "success", "data": { "token": "<jwt>", "user": { "id": 1, "username": "alice" } } }
```

### 认证前缀（需 JWT）
- 前缀：`/api/auth/*`
- 公共请求头：
```
Authorization: Bearer <token>
Content-Type: application/json    # 对于 POST
```

- 获取用户信息：`GET /api/auth/info/:user_id`
  - 响应体：
  ```json
  { "code": 0, "message": "success", "data": { "id": 1, "username": "alice", "email": "a@b.com" } }
  ```

- 更新资料：`POST /api/auth/update`
  - 请求体（任意字段可选）：
  ```json
  { "username": "alice2", "email": "a2@b.com", "avatar": "https://..." }
  ```
  - 响应体：
  ```json
  { "code": 0, "message": "success", "data": null }
  ```

- 修改密码：`POST /api/auth/password`
  - 请求体：
  ```json
  { "old_password": "old", "new_password": "new-123456" }
  ```
  - 响应体：
  ```json
  { "code": 0, "message": "success", "data": null }
  ```

## 内容（content）
- 健康检查: `GET /health`
- 公共请求头：
```
Content-Type: application/json   # 仅对 POST 场景，一般 GET 无需
```

- 获取文章：`GET /api/article/:article_id`
  - 响应体：
  ```json
  { "code": 0, "message": "success", "data": { "id": 1, "title": "T", "content": "..." } }
  ```

- 文章摘要列表：`GET /api/article/list?page=&page_size=`
  - 响应体：
  ```json
  { "code": 0, "message": "success", "data": { "list": [{"id":1,"title":"T"}], "total": 100, "page": 1, "page_size": 10 } }
  ```

- 关键词搜索：`GET /api/article/search?q=&page=&page_size=`
  - 响应体：同上，`list` 为匹配到的摘要列表

- 分类树（三级）：`GET /api/category/tree`
  - 响应体：
  ```json
  { "code": 0, "message": "success", "data": [{"id":1,"name":"A","children":[{"id":2,"name":"B","children":[]}]}] }
  ```

## 管理（admin）
- 健康检查: `GET /health`
- 登录：`POST /api/admin/login`
  - 请求头：`Content-Type: application/json`
  - 请求体：
  ```json
  { "username": "admin", "password": "secret" }
  ```
  - 响应体：
  ```json
  { "code": 0, "message": "success", "data": { "token": "mock-admin-token" } }
  ```

- 用户管理：
  - 分页：`GET /api/admin/users?page=&page_size=`
    - 响应体：`{ code,message,data:{ list:[], total:0 } }`
  - 新增：`POST /api/admin/users`
    - 请求头：`Content-Type: application/json`
    - 请求体：
    ```json
    { "username":"u","email":"u@a.com","password":"p","role":"admin","avatar":"" }
    ```
  - 修改：`POST /api/admin/users/update/:id`
    - 请求体（示例）：`{"email":"u2@a.com","role":"user"}`
  - 删除：`POST /api/admin/users/delete/:id`
    - 请求体：空

- 文章管理：
  - 分页：`GET /api/admin/articles?page=&page_size=`
  - 新增：`POST /api/admin/articles`
    - 请求头：`Content-Type: application/json`
    - 请求体（示例）：
    ```json
    { "title":"T","slug":"t","content":"...","summary":"...","author_id":1,"category_id":2,"status":0,"is_top":false,"is_recommend":false }
    ```
  - 修改：`POST /api/admin/articles/update/:id`
    - 请求体（示例）：`{"title":"T2","category_id":3}`
  - 删除：`POST /api/admin/articles/delete/:id`
    - 请求体：空

- 分类管理：
  - 分页：`GET /api/admin/categories?page=&page_size=`
  - 新增：`POST /api/admin/categories`
    - 请求头：`Content-Type: application/json`
    - 请求体（示例）：`{"name":"后端","slug":"backend","parent_id":0,"sort":10}`
  - 修改：`POST /api/admin/categories/update/:id`
    - 请求体（示例）：`{"name":"服务端","sort":20}`
  - 删除：`POST /api/admin/categories/delete/:id`
    - 请求体：空

## 统计（stat）
- 健康检查: `GET /health`

- 自增统计：`POST /api/stat/incr?type=view&target_id=1&target_type=article[&user_id=2]`
  - 请求头：
  ```
  Content-Type: application/json
  ```
  - 请求体：空（参数在 query）
  - 响应体：
  ```json
  { "code": 0, "message": "success", "data": null }
  ```

- 获取统计：`GET /api/stat/get?type=view&target_id=1&target_type=article[&user_id=2]`
  - 响应体：
  ```json
  { "code": 0, "message": "success", "data": { "value": 123 } }
  ```

---

# 十条典型用例

1. 登录：POST /api/login -> 返回 token
2. 获取用户信息：GET /api/auth/info/1 -> 返回 user
3. 文章详情：GET /api/article/1 -> 返回 article
4. 文章摘要列表：GET /api/article/list?page=1&page_size=10 -> 返回 list,total
5. 关键词搜索：GET /api/article/search?q=go&page=1&page_size=10 -> 返回 list,total
6. 分类树：GET /api/category/tree -> 返回三级分类树
7. 管理端新增文章：POST /api/admin/articles -> success
8. 管理端更新文章：POST /api/admin/articles/update/1 -> success
9. 管理端删除文章：POST /api/admin/articles/delete/1 -> success
10. 点赞记录：POST /api/stat/incr?type=like&target_id=1&target_type=article&user_id=2 -> success
