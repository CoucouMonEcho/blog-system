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
- 代理入口: `GET|POST /api/*`（自动路由到对应服务）

## 用户（user，前缀 /api/user）
- 健康检查: `GET /health`
- 登录: `POST /api/login` {username,password}
- 认证路由前缀: `/api/auth/*`（需 Authorization: Bearer <token>）
  - 获取用户信息: `GET /api/auth/info/:user_id`
  - 更新资料: `POST /api/auth/update` {username?,email?,avatar?}
  - 修改密码: `POST /api/auth/password` {old_password,new_password}

## 内容（content，前缀 /api/content）
- 健康检查: `GET /health`
- 获取文章: `GET /api/article/:article_id`
- 文章摘要列表: `GET /api/article/list?page=&page_size=`
- 关键词搜索: `GET /api/article/search?q=&page=&page_size=`
- 分类树（三级）: `GET /api/category/tree`

## 评论
已移除

## 统计（stat，前缀 /api/stat）
- 健康检查: `GET /health`
- 自增统计: `POST /api/stat/incr?type=view&target_id=1&target_type=article[&user_id=2]`
- 获取统计: `GET /api/stat/get?type=view&target_id=1&target_type=article[&user_id=2]`

## 管理（admin，前缀 /api/admin）
- 健康检查: `GET /health`
- 登录（示例）: `POST /api/admin/login` {username,password}
- 用户管理：
  - 分页：`GET /api/admin/users?page=&page_size=`
  - 新增：`POST /api/admin/users` {username,email,password,role,avatar?}
  - 修改：`POST /api/admin/users/update/:id`
  - 删除：`POST /api/admin/users/delete/:id`
- 文章管理：
  - 分页：`GET /api/admin/articles?page=&page_size=`
  - 新增：`POST /api/admin/articles` {title,slug,content,summary?,author_id,category_id,status,is_top?,is_recommend?}
  - 修改：`POST /api/admin/articles/update/:id`
  - 删除：`POST /api/admin/articles/delete/:id`
- 分类管理：
  - 分页：`GET /api/admin/categories?page=&page_size=`
  - 新增：`POST /api/admin/categories` {name,slug,parent_id?,sort?}
  - 修改：`POST /api/admin/categories/update/:id`
  - 删除：`POST /api/admin/categories/delete/:id`

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
