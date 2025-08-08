#!/bin/bash

# Blog System 轻量级部署脚本
set -e

# 配置变量
DEPLOY_PATH="/opt/blog-system"
SERVICE_NAME="user-service"
GATEWAY_SERVICE_NAME="gateway-service"
LOG_PATH="/var/log/blog-system"

# 日志函数
log_info() {
    echo "[INFO] $1"
}

log_error() {
    echo "[ERROR] $1"
}

# 部署单个服务
deploy_service() {
    local service_name=$1
    local service_path=$2
    local config_file=$3
    
    log_info "开始部署 $service_name..."
    
    # 停止现有服务
    log_info "停止现有服务..."
    systemctl stop ${service_name} >/dev/null 2>&1 || true
    
    # 更新配置文件
    log_info "更新配置文件..."
    if [ ! -z "$BLOG_PASSWORD" ]; then
        sed -i "s/BLOG_PASSWORD/$BLOG_PASSWORD/g" ${DEPLOY_PATH}/configs/${config_file}
        log_info "数据库密码已更新"
    fi
    
    # 修复日志路径 - 使用相对路径避免权限问题
    sed -i "s|logs/.*\.log|${DEPLOY_PATH}/logs/${service_name}.log|g" ${DEPLOY_PATH}/configs/${config_file}
    log_info "日志路径已更新为: ${DEPLOY_PATH}/logs/${service_name}.log"
    
    # 构建应用
    log_info "构建应用..."
    cd ${DEPLOY_PATH}/${service_path}
    
    export GOOS=linux
    export GOARCH=amd64
    export CGO_ENABLED=0
    
    go mod download
    go build -ldflags="-s -w" -o ${service_name} .
    
    if [ ! -f "${service_name}" ]; then
        log_error "应用构建失败"
        exit 1
    fi
    log_info "应用构建成功"
    
    # 创建systemd服务文件
    log_info "创建systemd服务文件..."
    
    cat > /etc/systemd/system/${service_name}.service << EOF
[Unit]
Description=Blog System ${service_name}
After=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=${DEPLOY_PATH}/${service_path}
ExecStart=${DEPLOY_PATH}/${service_path}/${service_name}
Restart=always
RestartSec=3
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    # 设置权限
    chmod +x ${DEPLOY_PATH}/${service_path}/${service_name}
    chmod 644 /etc/systemd/system/${service_name}.service
    
    # 重新加载systemd并启动服务
    systemctl daemon-reload
    systemctl enable ${service_name}
    systemctl start ${service_name}
    
    # 等待服务启动
    sleep 8
    
    # 检查服务状态
    if systemctl is-active --quiet ${service_name}; then
        log_info "${service_name} 启动成功"
    else
        log_error "${service_name} 启动失败"
        echo "=== 服务状态 ==="
        systemctl status ${service_name} --no-pager -l
        echo "=== 服务日志 ==="
        journalctl -u ${service_name} --no-pager -l
        echo "=== 配置文件内容 ==="
        cat ${DEPLOY_PATH}/configs/${config_file}
        echo "=== 日志文件 ==="
        ls -la ${DEPLOY_PATH}/logs/ || echo "日志目录不存在"
        exit 1
    fi
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
    
    # 检查Redis Cluster配置
    log_info "检查Redis Cluster配置..."
    if ! command -v redis-cli &> /dev/null; then
        log_error "Redis客户端未安装，请先安装Redis"
        exit 1
    fi
    
    # 检查Redis Cluster节点
    for port in 7001 7002 7003; do
        if ! redis-cli -p $port ping &> /dev/null; then
            log_error "Redis Cluster节点 $port 不可用"
            log_info "请确保Redis Cluster已正确配置并运行在端口 7001, 7002, 7003"
        else
            log_info "Redis Cluster节点 $port 正常"
        fi
    done
    
    # 部署用户服务
    deploy_service "user-service" "services/user" "user.yaml"
    
    # 检查用户服务端口
    if netstat -tlnp | grep :8001 >/dev/null; then
        log_info "用户服务端口8001监听正常"
    else
        log_error "用户服务端口8001未监听"
        exit 1
    fi
    
    # 部署网关服务
    deploy_service "gateway-service" "services/gateway" "gateway.yaml"
    
    # 检查网关服务端口
    if netstat -tlnp | grep :8000 >/dev/null; then
        log_info "网关服务端口8000监听正常"
    else
        log_error "网关服务端口8000未监听"
        exit 1
    fi
    
    log_info "部署完成！"
    log_info "用户服务地址: http://$(hostname -I | awk '{print $1}'):8001"
    log_info "网关服务地址: http://$(hostname -I | awk '{print $1}'):8000"
    log_info "日志文件: ${DEPLOY_PATH}/logs/"
}

# 执行主函数
main 