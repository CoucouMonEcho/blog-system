#!/bin/bash

# Blog System 轻量级部署脚本
# 适用于轻量级服务器，无需Docker

set -e

# 配置变量
APP_NAME="blog-system"
DEPLOY_PATH="/opt/${APP_NAME}"
SERVICE_NAME="user-service"
LOG_PATH="/var/log/${APP_NAME}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查系统要求
check_requirements() {
    log_info "检查系统要求..."
    
    # 检查Go是否安装
    if ! command -v go &> /dev/null; then
        log_error "Go未安装，请先安装Go 1.24.2+"
        exit 1
    fi
    
    # 检查Go版本
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go版本: $GO_VERSION"
    
    # 检查MySQL是否运行
    if ! systemctl is-active --quiet mysqld; then
        log_warn "MySQL服务未运行，请确保MySQL已安装并启动"
        log_info "尝试启动MySQL服务..."
        systemctl start mysqld || log_warn "MySQL启动失败"
    else
        log_info "MySQL服务运行正常"
    fi
    
    # 检查Redis是否运行
    if ! systemctl is-active --quiet redis-cli -p 7001; then
        log_warn "Redis服务未运行，请确保Redis已安装并启动"
        log_info "尝试启动Redis服务..."
        systemctl start redis-cli -p 7001 || log_warn "Redis启动失败"
    else
        log_info "Redis服务运行正常"
    fi
}

# 创建目录结构
create_directories() {
    log_info "创建目录结构..."
    
    mkdir -p ${DEPLOY_PATH}
    mkdir -p ${LOG_PATH}
    mkdir -p ${DEPLOY_PATH}/configs
    mkdir -p ${DEPLOY_PATH}/services
    mkdir -p ${DEPLOY_PATH}/scripts
    mkdir -p ${DEPLOY_PATH}/common
    
    log_info "目录创建完成"
}

# 停止现有服务
stop_service() {
    log_info "停止现有服务..."
    
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        systemctl stop ${SERVICE_NAME}
        log_info "服务已停止"
    else
        log_info "服务未运行"
    fi
}

# 备份现有配置
backup_config() {
    log_info "备份现有配置..."
    
    if [ -d "${DEPLOY_PATH}/configs" ]; then
        cp -r ${DEPLOY_PATH}/configs ${DEPLOY_PATH}/configs.backup.$(date +%Y%m%d_%H%M%S)
        log_info "配置备份完成"
    fi
}

# 更新配置文件
update_config() {
    log_info "更新配置文件..."
    
    # 替换数据库密码
    if [ ! -z "$BLOG_PASSWORD" ]; then
        sed -i "s/BLOG_PASSWORD/$BLOG_PASSWORD/g" ${DEPLOY_PATH}/configs/user.yaml
        log_info "数据库密码已更新"
    else
        log_warn "BLOG_PASSWORD环境变量未设置"
    fi
    
    # 设置日志路径
    sed -i "s|logs/user.log|${LOG_PATH}/user.log|g" ${DEPLOY_PATH}/configs/user.yaml
}

# 构建应用
build_application() {
    log_info "构建应用..."
    
    cd ${DEPLOY_PATH}/services/user
    
    # 设置Go环境变量
    export GOOS=linux
    export GOARCH=amd64
    export CGO_ENABLED=0
    
    # 下载依赖
    go mod download
    
    # 构建应用（静态链接，避免GLIBC版本问题）
    go build -ldflags="-s -w" -o user-service .
    
    if [ $? -eq 0 ]; then
        log_info "应用构建成功"
        
        # 验证二进制文件
        if [ -f "user-service" ]; then
            log_info "二进制文件大小: $(ls -lh user-service | awk '{print $5}')"
            log_info "二进制文件类型: $(file user-service)"
        else
            log_error "二进制文件未生成"
            exit 1
        fi
    else
        log_error "应用构建失败"
        exit 1
    fi
}

# 创建systemd服务文件
create_service_file() {
    log_info "创建systemd服务文件..."
    
    cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Blog System User Service
After=network.target mysql.service redis.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=${DEPLOY_PATH}/services/user
ExecStart=${DEPLOY_PATH}/services/user/user-service
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
Environment=GOMAXPROCS=1

[Install]
WantedBy=multi-user.target
EOF

    # 重新加载systemd
    systemctl daemon-reload
    log_info "服务文件创建完成"
}

# 启动服务
start_service() {
    log_info "启动服务..."
    
    # 设置文件权限
    chown -R www-data:www-data ${DEPLOY_PATH}
    chmod +x ${DEPLOY_PATH}/services/user/user-service
    
    # 启用服务
    systemctl enable ${SERVICE_NAME}
    
    # 启动服务
    systemctl start ${SERVICE_NAME}
    
    # 检查服务状态
    sleep 5
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        log_info "服务启动成功"
    else
        log_error "服务启动失败"
        systemctl status ${SERVICE_NAME} --no-pager -l
        log_error "查看详细日志: journalctl -u ${SERVICE_NAME} -f"
        exit 1
    fi
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 等待服务启动
    sleep 10
    
    # 检查端口是否监听
    if netstat -tlnp | grep :8001 > /dev/null; then
        log_info "服务端口8001监听正常"
    else
        log_error "服务端口8001未监听"
        log_error "检查服务状态: systemctl status ${SERVICE_NAME}"
        log_error "查看服务日志: journalctl -u ${SERVICE_NAME} -f"
        exit 1
    fi
    
    # 测试API接口
    if curl -f http://localhost:8001/health > /dev/null 2>&1; then
        log_info "API健康检查通过"
    else
        log_warn "API健康检查失败，但服务已启动"
        log_info "尝试访问: curl http://localhost:8001/health"
    fi
}

# 显示服务状态
show_status() {
    log_info "显示服务状态..."
    
    echo "=== 服务状态 ==="
    systemctl status ${SERVICE_NAME} --no-pager -l
    
    echo "=== 端口监听 ==="
    netstat -tlnp | grep :8001 || echo "端口8001未监听"
    
    echo "=== 日志文件 ==="
    ls -la ${LOG_PATH}/ || echo "日志目录不存在"
    
    echo "=== 进程信息 ==="
    ps aux | grep user-service | grep -v grep || echo "进程未运行"
    
    echo "=== 二进制文件 ==="
    ls -la ${DEPLOY_PATH}/services/user/user-service || echo "二进制文件不存在"
}

# 主函数
main() {
    log_info "开始部署 Blog System..."
    
    # 检查是否为root用户
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用root权限运行此脚本"
        exit 1
    fi
    
    # 检查环境变量
    if [ -z "$BLOG_PASSWORD" ]; then
        log_warn "BLOG_PASSWORD环境变量未设置，将使用配置文件中的默认值"
    fi
    
    check_requirements
    create_directories
    stop_service
    backup_config
    update_config
    build_application
    create_service_file
    start_service
    health_check
    show_status
    
    log_info "部署完成！"
    log_info "服务地址: http://$(hostname -I | awk '{print $1}'):8001"
    log_info "日志文件: ${LOG_PATH}/user.log"
    log_info "配置文件: ${DEPLOY_PATH}/configs/user.yaml"
}

# 脚本入口
if [ "$1" = "status" ]; then
    show_status
elif [ "$1" = "stop" ]; then
    stop_service
elif [ "$1" = "start" ]; then
    start_service
elif [ "$1" = "restart" ]; then
    stop_service
    start_service
else
    main
fi 