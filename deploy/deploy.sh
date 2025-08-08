#!/bin/bash

# Blog System 轻量级部署脚本
set -e

# 配置变量
DEPLOY_PATH="/opt/blog-system"
SERVICE_NAME="user-service"
LOG_PATH="/var/log/blog-system"

# 日志函数
log_info() {
    echo "[INFO] $1"
}

log_error() {
    echo "[ERROR] $1"
}

# 主函数
main() {
    log_info "开始部署 Blog System..."
    
    # 检查是否为root用户
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用root权限运行此脚本"
        exit 1
    fi
    
    # 创建目录结构
    log_info "创建目录结构..."
    mkdir -p ${LOG_PATH}
    mkdir -p ${DEPLOY_PATH}/logs
    
    # 停止现有服务
    log_info "停止现有服务..."
    systemctl stop ${SERVICE_NAME} >/dev/null 2>&1 || true
    
    # 更新配置文件
    log_info "更新配置文件..."
    if [ ! -z "$BLOG_PASSWORD" ]; then
        sed -i "s/BLOG_PASSWORD/$BLOG_PASSWORD/g" ${DEPLOY_PATH}/configs/user.yaml
        log_info "数据库密码已更新"
    fi
    
    # 修复日志路径 - 使用相对路径避免权限问题
    sed -i "s|logs/user.log|${DEPLOY_PATH}/logs/user.log|g" ${DEPLOY_PATH}/configs/user.yaml
    log_info "日志路径已更新为: ${DEPLOY_PATH}/logs/user.log"
    
    # 构建应用
    log_info "构建应用..."
    cd ${DEPLOY_PATH}/services/user
    
    export GOOS=linux
    export GOARCH=amd64
    export CGO_ENABLED=0
    
    go mod download
    go build -ldflags="-s -w" -o user-service .
    
    if [ ! -f "user-service" ]; then
        log_error "应用构建失败"
        exit 1
    fi
    log_info "应用构建成功"
    
    # 创建systemd服务文件
    log_info "创建systemd服务文件..."
    
    cat > /etc/systemd/system/${SERVICE_NAME}.service << 'EOF'
[Unit]
Description=Blog System User Service
After=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/opt/blog-system/services/user
ExecStart=/opt/blog-system/services/user/user-service
Restart=always
RestartSec=3
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    # 设置权限
    chmod +x ${DEPLOY_PATH}/services/user/user-service
    chmod 644 /etc/systemd/system/${SERVICE_NAME}.service
    
    # 重新加载systemd并启动服务
    systemctl daemon-reload
    systemctl enable ${SERVICE_NAME}
    systemctl start ${SERVICE_NAME}
    
    # 等待服务启动
    sleep 8
    
    # 检查服务状态
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        log_info "服务启动成功"
    else
        log_error "服务启动失败"
        echo "=== 服务状态 ==="
        systemctl status ${SERVICE_NAME} --no-pager -l
        echo "=== 服务日志 ==="
        journalctl -u ${SERVICE_NAME} --no-pager -l
        echo "=== 配置文件内容 ==="
        cat ${DEPLOY_PATH}/configs/user.yaml
        echo "=== 日志文件 ==="
        ls -la ${DEPLOY_PATH}/logs/ || echo "日志目录不存在"
        exit 1
    fi
    
    # 检查端口
    if netstat -tlnp | grep :8001 >/dev/null; then
        log_info "端口8001监听正常"
    else
        log_error "端口8001未监听"
        exit 1
    fi
    
    log_info "部署完成！"
    log_info "服务地址: http://$(hostname -I | awk '{print $1}'):8001"
    log_info "日志文件: ${DEPLOY_PATH}/logs/user.log"
}

# 执行主函数
main 