# Blog System 轻量级部署指南

本文档详细说明如何通过 GitHub Actions 将 Blog System 部署到轻量级云服务器。

## 📋 前置要求

### 1. 云服务器准备
- 一台运行 Linux 的轻量级云服务器（推荐 Ubuntu 20.04+）
- 服务器已安装 Go 1.24.2+
- 服务器已安装 MySQL 8.0+
- 服务器已安装 Redis 7.0+
- 服务器已配置 SSH 密钥认证

### 2. GitHub 仓库设置
- 项目已推送到 GitHub 仓库
- 仓库已启用 GitHub Actions

## 🔧 GitHub 仓库配置

### 1. 设置 Secrets

在 GitHub 仓库中设置以下 Secrets：

#### SSH 连接配置
```
SSH_HOST          # 服务器IP地址
SSH_USERNAME      # SSH用户名（如：root）
SSH_PRIVATE_KEY   # SSH私钥内容
```

#### 数据库配置
```
BLOG_PASSWORD     # 数据库密码
```

### 2. 设置方法

1. 进入 GitHub 仓库页面
2. 点击 `Settings` → `Secrets and variables` → `Actions`
3. 点击 `New repository secret`
4. 添加上述每个 Secret

## 📁 项目结构

确保项目包含以下文件：

```
blog-system/
├── .github/
│   └── workflows/
│       └── deploy.yml
├── common/
│   └── pkg/
│       └── logger/
│           └── logger.go
├── configs/
│   └── user.yaml
├── deploy/
│   ├── README.md
│   └── deploy.sh
├── services/
│   └── user/
└── README.md
```

## 🚀 GitHub Actions 工作流

### 1. 创建工作流文件

在 `.github/workflows/deploy.yml` 中创建以下内容：

```yaml
name: Deploy to Production

on:
  push:
    branches: [ main ]
  workflow_dispatch:

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
      
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.24.2'
        
    - name: Build application
      run: |
        cd services/user
        go mod download
        go build -o user-service .
        
    - name: Deploy to server
      uses: appleboy/ssh-action@v0.1.5
      with:
        host: ${{ secrets.SSH_HOST }}
        username: ${{ secrets.SSH_USERNAME }}
        key: ${{ secrets.SSH_PRIVATE_KEY }}
        port: 22
        script: |
          mkdir -p /opt/blog-system
          systemctl stop user-service || true
          rm -rf /opt/blog-system/services /opt/blog-system/configs /opt/blog-system/deploy
          
    - name: Upload files
      uses: appleboy/scp-action@v0.1.4
      with:
        host: ${{ secrets.SSH_HOST }}
        username: ${{ secrets.SSH_USERNAME }}
        key: ${{ secrets.SSH_PRIVATE_KEY }}
        source: "services,configs,deploy"
        target: /opt/blog-system
        
    - name: Execute deployment script
      uses: appleboy/ssh-action@v0.1.5
      with:
        host: ${{ secrets.SSH_HOST }}
        username: ${{ secrets.SSH_USERNAME }}
        key: ${{ secrets.SSH_PRIVATE_KEY }}
        envs: BLOG_PASSWORD
        script: |
          cd /opt/blog-system
          export BLOG_PASSWORD="${{ secrets.BLOG_PASSWORD }}"
          chmod +x deploy/deploy.sh
          ./deploy/deploy.sh
```

## 🐧 服务器环境准备

### 1. 安装 Go

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# 或者下载最新版本
wget https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

### 2. 安装 MySQL

```bash
# Ubuntu/Debian
sudo apt install mysql-server
sudo systemctl start mysql
sudo systemctl enable mysql

# 配置MySQL
sudo mysql_secure_installation

# 创建数据库和用户
sudo mysql -u root -p
CREATE DATABASE blog_user;
CREATE USER 'blog_user'@'localhost' IDENTIFIED BY 'your_secure_password';
GRANT ALL PRIVILEGES ON blog_user.* TO 'blog_user'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### 3. 安装 Redis

```bash
# Ubuntu/Debian
sudo apt install redis-server
sudo systemctl start redis
sudo systemctl enable redis

# 验证Redis
redis-cli ping
```

### 4. 创建系统用户

```bash
# 创建应用用户
sudo useradd -r -s /bin/false www-data
sudo usermod -aG www-data www-data
```

## 🔐 SSH 密钥配置

### 1. 生成 SSH 密钥对

```bash
# 在本地生成密钥对
ssh-keygen -t rsa -b 4096 -C "github-actions"

# 查看公钥
cat ~/.ssh/id_rsa.pub

# 查看私钥
cat ~/.ssh/id_rsa
```

### 2. 配置服务器

```bash
# 在服务器上添加公钥
echo "你的公钥内容" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
chmod 700 ~/.ssh

# 测试SSH连接
ssh username@your-server-ip
```

### 3. 配置 GitHub Secrets

将私钥内容复制到 `SSH_PRIVATE_KEY` Secret 中。

## 📝 配置文件管理

### 1. 更新 user.yaml

确保 `configs/user.yaml` 中的配置适合轻量级部署：

```yaml
app:
  name: user-service
  port: 8001

database:
  driver: "mysql"
  host: localhost  # 本地MySQL
  port: 3306
  user: root
  password: BLOG_PASSWORD  # 将被替换
  name: blog_user

redis:
  addr: localhost:6379  # 本地Redis
  password: ""

log:
  level: info
  path: /var/log/blog-system/user.log
```

## 🚀 部署步骤

### 1. 手动触发部署

1. 进入 GitHub 仓库
2. 点击 `Actions` 标签
3. 选择 `Deploy to Production` 工作流
4. 点击 `Run workflow`
5. 选择分支并点击 `Run workflow`

### 2. 自动部署

推送代码到 `main` 分支将自动触发部署。

## 🔍 部署验证

### 1. 检查服务状态

```bash
# SSH 到服务器
ssh username@your-server-ip

# 检查服务状态
systemctl status user-service

# 查看服务日志
journalctl -u user-service -f
```

### 2. 测试 API

```bash
# 测试服务健康状态
curl http://your-server-ip:8001/health

# 测试用户注册
curl -X POST http://your-server-ip:8001/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"123456"}'

# 测试用户登录
curl -X POST http://your-server-ip:8001/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'
```

## 🛠️ 故障排除

### 1. 常见问题

#### SSH 连接失败
- 检查 `SSH_HOST` 和 `SSH_USERNAME` 是否正确
- 确认服务器防火墙允许 SSH 连接
- 验证 SSH 密钥是否正确配置

#### 数据库连接失败
- 检查 `BLOG_PASSWORD` Secret 是否正确设置
- 确认 MySQL 服务是否正常启动
- 检查数据库用户权限

#### 服务启动失败
- 查看服务日志：`journalctl -u user-service -f`
- 检查端口是否被占用：`netstat -tlnp | grep 8001`
- 确认配置文件格式正确

### 2. 日志查看

```bash
# 查看应用日志
tail -f /var/log/blog-system/user.log

# 查看系统服务日志
journalctl -u user-service -f

# 查看实时日志
journalctl -u user-service -f --since "5 minutes ago"
```

## 📊 监控和维护

### 1. 服务监控

```bash
# 查看服务状态
systemctl status user-service

# 查看资源使用情况
ps aux | grep user-service

# 查看日志
journalctl -u user-service -f
```

### 2. 备份策略

```bash
# 备份数据库
mysqldump -u root -p blog_user > backup_$(date +%Y%m%d_%H%M%S).sql

# 备份配置文件
cp -r /opt/blog-system/configs /backup/configs_$(date +%Y%m%d_%H%M%S)

# 备份日志
cp /var/log/blog-system/user.log /backup/user_$(date +%Y%m%d_%H%M%S).log
```

## 🔄 更新部署

### 1. 自动更新

推送代码到 `main` 分支将自动触发重新部署。

### 2. 手动更新

```bash
# SSH 到服务器
ssh username@your-server-ip

# 进入部署目录
cd /opt/blog-system

# 拉取最新代码
git pull origin main

# 重新部署
./deploy/deploy.sh
```

### 3. 服务管理

```bash
# 停止服务
sudo systemctl stop user-service

# 启动服务
sudo systemctl start user-service

# 重启服务
sudo systemctl restart user-service

# 查看状态
sudo systemctl status user-service

# 启用开机自启
sudo systemctl enable user-service
```

## 📈 性能优化

### 1. 系统优化

```bash
# 调整文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# 调整内核参数
echo "net.core.somaxconn = 65535" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65535" >> /etc/sysctl.conf
sysctl -p
```

### 2. MySQL 优化

```bash
# 编辑MySQL配置
sudo nano /etc/mysql/mysql.conf.d/mysqld.cnf

# 添加以下配置
[mysqld]
innodb_buffer_pool_size = 256M
innodb_log_file_size = 64M
max_connections = 200
```

### 3. Redis 优化

```bash
# 编辑Redis配置
sudo nano /etc/redis/redis.conf

# 添加以下配置
maxmemory 256mb
maxmemory-policy allkeys-lru
```

## 🔒 安全加固

### 1. 防火墙配置

```bash
# 安装ufw
sudo apt install ufw

# 配置防火墙规则
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 8001
sudo ufw enable
```

### 2. 系统安全

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装安全工具
sudo apt install fail2ban

# 配置fail2ban
sudo nano /etc/fail2ban/jail.local
```

### 3. 数据库安全

```bash
# 删除匿名用户
sudo mysql -u root -p
DELETE FROM mysql.user WHERE User='';
FLUSH PRIVILEGES;
EXIT;

# 限制远程访问
sudo nano /etc/mysql/mysql.conf.d/mysqld.cnf
# 添加 bind-address = 127.0.0.1
```

## 📞 技术支持

如遇到部署问题，请检查：

1. **GitHub Actions 日志**: 查看构建和部署日志
2. **服务器系统日志**: `journalctl -u user-service`
3. **应用日志文件**: `/var/log/blog-system/user.log`
4. **网络连接状态**: `netstat -tlnp | grep 8001`

### 常见错误及解决方案

#### 1. 权限错误
```bash
# 修复文件权限
sudo chown -R www-data:www-data /opt/blog-system
sudo chmod +x /opt/blog-system/deploy/deploy.sh
```

#### 2. 端口占用
```bash
# 查看端口占用
sudo netstat -tlnp | grep 8001

# 杀死占用进程
sudo kill -9 <PID>
```

#### 3. 数据库连接失败
```bash
# 检查MySQL状态
sudo systemctl status mysql

# 重启MySQL
sudo systemctl restart mysql

# 检查连接
mysql -u root -p -h localhost
```

## 💡 轻量级部署优势

### 1. 资源占用少
- 无需 Docker 容器开销
- 直接使用系统服务
- 内存占用更少

### 2. 部署简单
- 无需容器编排
- 直接使用 shell 脚本
- 配置更直观

### 3. 维护方便
- 使用 systemd 管理服务
- 日志集成到系统日志
- 监控更简单

### 4. 成本效益
- 适合轻量级服务器
- 资源利用率高
- 运维成本低

---

**注意**: 请确保所有敏感信息（如数据库密码）都通过 GitHub Secrets 管理，不要直接写在配置文件中。 