# Blog System - 基于DDD的博客系统

基于 DDD（领域驱动设计）架构的现代化博客系统，使用 Go 语言和微服务架构构建，支持轻量级服务器部署。

## 🏗️ 项目架构

```
blog-system/
├── services/           # 微服务模块
│   ├── user/          # 用户服务 (端口: 8001)
│   ├── content/       # 内容服务 (端口: 8002)
│   ├── admin/         # 管理服务 (端口: 8003)
│   ├── stat/          # 统计服务 (端口: 8004)
│   └── gateway/       # 网关服务 (端口: 8000)
├── common/            # 通用库（日志/DTO/错误码/工具）
├── configs/           # 各服务配置（user/content/stat/admin/gateway）
├── deploy/            # 部署脚本与文档
├── docs/              # API 文档
├── .github/workflows/ # CI/CD 工作流
└── go.work            # Go 工作区配置
```

## 🛠️ 技术栈

### 核心框架
- **[go-framework](https://github.com/CoucouMonEcho/go-framework)**
  - `web`: HTTP 框架 + recover/otel/prometheus 中间件
  - `orm`: ORM + otel/prometheus 中间件
  - `cache`: Redis 封装（Cluster）
  - `micro/registry`: etcd 服务注册

### 数据存储
- **MySQL 8.0**: 主数据库
- **Redis 7**: 缓存和会话存储

### 部署和运维
- **轻量级部署**: Shell 脚本 + systemd
- **GitHub Actions**: 自动化 CI/CD
- **结构化日志**: 多级路径区分

### 架构模式
- **DDD (领域驱动设计)**: 分层架构
- **微服务**: 服务拆分和独立部署
- **CQRS**: 命令查询职责分离（规划中）

## 🚀 快速开始

### 1. 环境要求

- Go 1.24.2+
- MySQL 8.0+
- Redis 7.0+
- Linux 服务器（推荐 Ubuntu 20.04+）

### 2. 克隆项目

```bash
git clone <repository-url>
cd blog-system
```

### 3. 本地开发（示例）

```bash
# 安装依赖
go work sync

# 启动用户服务
cd services/user
go run main.go
```

### 4. 生产环境部署

```bash
# 使用 GitHub Actions 自动部署
# 推送代码到 main 分支即可触发部署

# 或手动部署
chmod +x deploy/deploy.sh
./deploy/deploy.sh
```

## 📋 服务说明

### ✅ 用户服务 (user)
- **功能**: 登录、信息查询/更新（注册已迁移至 admin）
- **端口**: 8001
- **特性**:
  - JWT 身份验证
  - 密码加密存储
  - Redis 缓存支持
  - 结构化日志
  - 轻量级部署

### ✅ 内容服务 (content)
- **功能**: 文章只读访问（详情、分页摘要列表、关键词搜索）、分类树查询（三级）
- **端口**: 8002
 - **说明**: 文章/分类的新增/修改/删除由 admin 负责

### ✅ 管理服务 (admin)
- **功能**: 用户管理（分页/增删改）、文章管理（分页/增删改）、分类管理（分页/增删改）
- **端口**: 8003
- **说明**: 负责用户注册、内容与分类的后台维护；更新分类后自动清理 `content:category_tree` 缓存

### ✅ 统计服务 (stat)
- **功能**: 浏览量、点赞统计
- **端口**: 8004
- **规划**: 热榜排行、数据可视化

### ✅ 网关服务 (gateway)
- **功能**: 统一入口、路由聚合
- **端口**: 8000
- **规划**: 负载均衡、限流熔断

## 🏛️ DDD 架构设计

### 分层架构

每个服务都遵循 DDD 分层架构：

```
service/
├── domain/           # 领域层
│   ├── entity.go     # 领域实体
│   └── repository.go # 仓储接口
├── application/      # 应用层
│   └── service.go    # 应用服务
├── infrastructure/   # 基础设施层
│   ├── repository.go # 仓储实现
│   ├── database.go   # 数据库连接
│   └── config.go     # 配置管理
├── api/              # 接口层
│   └── http.go       # HTTP 接口
└── main.go          # 服务入口
```

## 📡 API 文档

详见 `docs/api.md`（已与当前路由保持一致，去除评论模块、更新 content 与 admin 接口）。

## 🔧 配置管理

### 用户服务配置 (configs/user.yaml)

```yaml
app:
  name: user-service
  port: 8001

database:
  driver: "mysql"
  host: localhost
  port: 3306
  user: root
  password: BLOG_PASSWORD  # 环境变量注入
  name: blog_user

redis:
  addr: localhost:6379
  password: ""

log:
  level: info
  path: /var/log/blog-system/user.log
```

## 📊 日志系统

### 结构化日志

使用公共日志模块 `common/pkg/logger`，支持多级路径区分：

```
[2024-01-15 10:30:45.123][user-service][main][INFO] 服务启动成功
[2024-01-15 10:30:45.456][user-service][database][INFO] 数据库连接成功
[2024-01-15 10:30:45.789][user-service][cache][INFO] 缓存连接成功
```

## 🚀 部署架构

### 轻量级部署

- **systemd 服务管理**: 自动启动、重启、监控
- **Shell 脚本部署**: 简单、高效、易维护
- **GitHub Actions CI/CD**: 自动化构建和部署
- **结构化日志**: 便于问题排查和监控

### 部署流程

1. **代码推送**: 推送到 main 分支
2. **自动构建**: GitHub Actions 构建应用
3. **文件上传**: SSH 上传到服务器
4. **服务部署**: 执行部署脚本
5. **健康检查**: 验证服务状态

## 🔒 安全特性

### 身份验证
- **JWT 令牌**: 无状态身份验证
- **密码加密**: bcrypt 哈希存储
- **角色权限**: 基于角色的访问控制

### 数据安全
- **敏感信息**: 通过环境变量管理
- **数据库安全**: 最小权限原则
- **日志安全**: 不记录敏感信息

## 📈 监控和维护

### 服务监控
```bash
# 查看服务状态
systemctl status user-service

# 查看日志
journalctl -u user-service -f

# 查看端口监听
netstat -tlnp | grep :8001
```

### 性能监控
- **结构化日志**: 便于日志分析
- **健康检查**: 自动服务状态检测
- **资源监控**: 系统资源使用情况

## 🛠️ 开发指南

### 添加新服务

1. **创建服务目录**:
   ```bash
   mkdir -p services/new-service/{domain,application,infrastructure,api}
   ```

2. **创建 go.mod**:
   ```
   module blog-system/services/new-service
   go 1.24.2
   ```

3. **实现 DDD 分层**:
   - `domain/`: 领域实体和接口
   - `application/`: 应用服务
   - `infrastructure/`: 基础设施实现
   - `api/`: HTTP 接口

4. **更新 go.work**:
   ```
   use (
       ./services/new-service
   )
   ```

### 代码规范

- **DDD 分层**: 严格遵循分层架构
- **错误处理**: 统一错误码和错误处理
- **日志记录**: 使用结构化日志
- **配置管理**: 通过 YAML 配置文件

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系方式

- **项目维护者**: CoucouMonEcho
- **邮箱**: [coucoumonecho@gmail.com]
- **GitHub**: [https://github.com/CoucouMonEcho/blog-system]

---

**状态**: 🚧 开发中 - 用户服务已完成，其他服务正在开发中 