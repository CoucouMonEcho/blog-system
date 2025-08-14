# API 文档（Gateway / User / Content / Comment / Admin / Stat）

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
- 注册: `POST /api/register` {username,email,password}
- 登录: `POST /api/login` {username,password}
- 认证路由前缀: `/api/auth/*`（需 Authorization: Bearer <token>）
  - 获取用户信息: `GET /api/auth/info/:user_id`
  - 更新资料: `POST /api/auth/update` {username?,email?,avatar?}
  - 修改密码: `POST /api/auth/password` {old_password,new_password}

## 内容（content，前缀 /api/content）
- 健康检查: `GET /health`
- 创建文章: `POST /api/article` {title,slug,content,summary?,author_id,category_id,status}
- 获取文章: `GET /api/article/:id`
- 更新文章: `POST /api/article/update/:id` {title?,slug?,content?,summary?,category_id?,status?,is_top?,is_recommend?}
- 删除文章: `POST /api/article/delete/:id`

## 评论（comment，前缀 /api/comment）
- 健康检查: `GET /health`
- 发表评论: `POST /api/comment` {content,user_id,article_id,parent_id?}
- 获取评论: `GET /api/comment/:id`
- 按文章列出评论: `GET /api/comment/article/:article_id?page=1&page_size=10`
- 审核通过: `POST /api/comment/approve/:id`
- 审核拒绝: `POST /api/comment/reject/:id`
- 删除评论: `POST /api/comment/delete/:id`

## 统计（stat，前缀 /api/stat）
- 健康检查: `GET /health`
- 自增统计: `POST /api/stat/incr?type=view&target_id=1&target_type=article[&user_id=2]`
- 获取统计: `GET /api/stat/get?type=view&target_id=1&target_type=article[&user_id=2]`

## 管理（admin，前缀 /api/admin）
- 健康检查: `GET /health`
- 登录（示例）: `POST /api/admin/login` {username,password}

---

# 十条典型用例

1. 注册用户：POST /api/register -> 返回 user 与 token
2. 登录：POST /api/login -> 返回 token
3. 获取用户信息：GET /api/auth/info/1 -> 返回 user
4. 创建文章：POST /api/article -> 返回 article
5. 获取文章：GET /api/article/1 -> 返回 article
6. 更新文章：POST /api/article/update/1 -> 返回 success
7. 发布评论：POST /api/comment -> 返回 comment
8. 查看文章评论列表：GET /api/comment/article/1?page=1&page_size=10 -> 返回 list,total
9. 点赞记录：POST /api/stat/incr?type=like&target_id=1&target_type=article&user_id=2 -> success
10. 管理端登录：POST /api/admin/login -> 返回 token
