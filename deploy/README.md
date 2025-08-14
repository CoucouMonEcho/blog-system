# Blog System 部署文档

## 概述

本项目采用DDD微服务架构，使用统一的GitHub Actions工作流进行部署。

## 部署架构

### 服务依赖关系（概览）
```
Redis Cluster (7001, 7002, 7003)
    ↓
Common Module
    ├── User Service (8001)
    ├── Content Service (8002)
    └── Admin Service (8003)
    ├── Stat Service (8004)
         ↓
      Gateway Service (8000)
```

### 统一工作流

**deploy.yml** - 统一部署工作流（关键 job）：

1. **deploy-common** - 部署公共模块
   - 上传common模块到服务器
   - 验证部署成功

2. **deploy-user-service** - 依赖 `deploy-common`
3. **deploy-content-service** - 依赖 `deploy-common`（与 user 并行）
4. **deploy-stat-service** - 依赖 `deploy-common`
5. **deploy-admin-service** - 依赖 `deploy-common`
6. **deploy-gateway-service** - 依赖上述四个服务（确保全部部署成功后再发布网关）

## 部署脚本

### 主要脚本

1. **deploy.sh** - 完整部署脚本
   - 部署所有服务
   - 包含依赖检查
   - 包含服务验证

2. **deploy-user-service.sh** - 用户服务部署脚本
   - 只部署用户服务
   - 包含Redis依赖检查
   - 包含服务验证

3. **deploy-gateway-service.sh** - 网关服务部署脚本
   - 只部署网关服务
   - 包含用户服务依赖检查
   - 包含服务验证

4. **manage-deploy.sh** - 部署管理脚本
   - 统一管理所有部署操作
   - 支持服务状态查看
   - 支持服务重启/停止/更新

5. **test-deploy.sh** - 测试部署脚本
   - 用于验证部署过程
   - 测试构建、配置、服务启动

### 管理脚本用法

```bash
# 显示帮助
./manage-deploy.sh -h

# 列出所有服务
./manage-deploy.sh -l

# 显示所有服务状态
./manage-deploy.sh -s

# 显示特定服务状态
./manage-deploy.sh -s user

# 重启所有服务
./manage-deploy.sh -r all

# 重启特定服务
./manage-deploy.sh -r user

# 停止特定服务
./manage-deploy.sh -t gateway

# 更新特定服务
./manage-deploy.sh -u user
```

## 日志优化

### 问题解决

原问题：GitHub Actions日志中出现大量无用的 `err:` 和 `out:` 输出。

### 解决方案

1. **添加静默执行函数**
   ```bash
   silent_exec() {
       "$@" >/dev/null 2>&1
   }
   ```

2. **重定向所有命令输出**
   - 使用 `silent_exec` 包装所有不需要输出的命令
   - 保留重要的日志信息
   - 抑制构建过程和系统命令的冗余输出

3. **优化端口检查**
   - 使用更安静的方式检查端口监听
   - 避免管道命令产生不必要的输出

## 部署流程

### 自动部署

1. **推送代码到main分支**
   - 自动触发统一部署工作流
   - 按依赖顺序执行各个job

2. **手动触发**
   - 在GitHub Actions页面手动触发部署
   - 适用于紧急部署或测试

### 服务器管理

```bash
# 查看所有服务状态
./manage-deploy.sh -s

# 重启特定服务
./manage-deploy.sh -r user

# 更新特定服务
./manage-deploy.sh -u gateway

# 停止所有服务
./manage-deploy.sh -t all
```

## 验证部署

### 服务状态检查

```bash
# 检查所有服务状态
systemctl status user-service
systemctl status gateway-service

# 检查端口监听
netstat -tlnp | grep -E ":(8000|8001)"

# 检查服务日志
ls -la /opt/blog-system/logs/
```

### 服务访问测试

```bash
# 测试用户服务
curl http://localhost:8001/health

# 测试网关服务
curl http://localhost:8000/health
```

## 故障排除

### 常见问题

1. **服务启动失败**
   - 检查Redis Cluster是否正常运行
   - 检查配置文件是否正确
   - 查看服务日志：`journalctl -u user-service`

2. **端口未监听**
   - 检查服务是否正常启动
   - 检查防火墙设置
   - 检查端口是否被占用

3. **依赖服务未启动**
   - 确保按依赖顺序部署
   - 检查依赖服务状态
   - 重新启动依赖服务

4. **权限问题**
   - 确保脚本有执行权限：`chmod +x deploy/*.sh`
   - 检查文件权限设置

### 日志位置

- **服务日志**: `/opt/blog-system/logs/`
- **系统日志**: `journalctl -u <service-name>`
- **部署日志**: GitHub Actions 输出

## 安全注意事项

1. **权限管理**
   - 所有脚本需要root权限运行
   - 确保SSH密钥安全存储
   - 定期更新服务器密码

2. **网络安全**
   - 确保防火墙正确配置
   - 只开放必要的端口
   - 使用HTTPS进行安全通信

3. **数据安全**
   - 定期备份数据库
   - 加密敏感配置信息
   - 监控异常访问

## 性能优化

1. **构建优化**
   - 使用 `-ldflags="-s -w"` 减小二进制文件大小
   - 设置 `CGO_ENABLED=0` 提高兼容性
   - 使用多阶段构建减少镜像大小

2. **运行优化**
   - 配置合适的JVM参数
   - 优化数据库连接池
   - 使用CDN加速静态资源

3. **监控优化**
   - 配置服务监控
   - 设置告警机制
   - 定期检查性能指标

### 验证步骤

1. **构建测试**
   - 检查go.mod文件
   - 下载依赖
   - 构建应用

2. **配置测试**
   - 检查配置文件存在
   - 验证配置格式

3. **服务测试**
   - 创建systemd服务
   - 启动服务
   - 验证服务状态

## 更新日志

### v2.0.0 (当前版本)
- 统一部署工作流，使用3个job
- 优化日志输出，消除无用信息
- 修复权限问题
- 添加依赖检查机制
- 改进错误处理

### v1.0.0
- 初始部署脚本
- 基础服务部署功能 