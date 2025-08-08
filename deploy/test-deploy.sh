#!/bin/bash

# 测试部署脚本
set -e

# 配置变量
DEPLOY_PATH="/opt/blog-system"
SERVICE_NAME="user-service"

# 日志函数
log_info() {
    echo "[INFO] $1"
}

log_error() {
    echo "[ERROR] $1"
}

# 静默执行函数
silent_exec() {
    "$@" >/dev/null 2>&1
}

# 测试构建
test_build() {
    log_info "测试构建过程..."
    
    cd ${DEPLOY_PATH}/services/user
    
    export GOOS=linux
    export GOARCH=amd64
    export CGO_ENABLED=0
    
    # 检查go.mod
    if [ ! -f "go.mod" ]; then
        log_error "go.mod文件不存在"
        return 1
    fi
    
    # 下载依赖
    log_info "下载依赖..."
    silent_exec go mod download
    
    # 构建应用
    log_info "构建应用..."
    silent_exec go build -ldflags="-s -w" -o ${SERVICE_NAME} .
    
    if [ ! -f "${SERVICE_NAME}" ]; then
        log_error "应用构建失败"
        return 1
    fi
    
    log_info "应用构建成功"
    return 0
}

# 测试配置
test_config() {
    log_info "测试配置文件..."
    
    if [ ! -f "${DEPLOY_PATH}/configs/user.yaml" ]; then
        log_error "配置文件不存在"
        return 1
    fi
    
    log_info "配置文件存在"
    return 0
}

# 测试服务启动
test_service() {
    log_info "测试服务启动..."
    
    # 停止现有服务
    silent_exec systemctl stop ${SERVICE_NAME} || true
    
    # 创建服务文件
    cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Blog System ${SERVICE_NAME}
After=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=${DEPLOY_PATH}/services/user
ExecStart=${DEPLOY_PATH}/services/user/${SERVICE_NAME}
Restart=always
RestartSec=3
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    # 设置权限
    silent_exec chmod +x ${DEPLOY_PATH}/services/user/${SERVICE_NAME}
    silent_exec chmod 644 /etc/systemd/system/${SERVICE_NAME}.service
    
    # 重新加载systemd并启动服务
    silent_exec systemctl daemon-reload
    silent_exec systemctl enable ${SERVICE_NAME}
    silent_exec systemctl start ${SERVICE_NAME}
    
    # 等待服务启动
    sleep 5
    
    # 检查服务状态
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        log_info "服务启动成功"
        return 0
    else
        log_error "服务启动失败"
        systemctl status ${SERVICE_NAME} --no-pager -l
        return 1
    fi
}

# 主函数
main() {
    log_info "开始测试部署..."
    
    # 检查是否为root用户
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用root权限运行此脚本"
        exit 1
    fi
    
    # 创建目录结构
    silent_exec mkdir -p ${DEPLOY_PATH}/logs
    
    # 测试构建
    if ! test_build; then
        log_error "构建测试失败"
        exit 1
    fi
    
    # 测试配置
    if ! test_config; then
        log_error "配置测试失败"
        exit 1
    fi
    
    # 测试服务启动
    if ! test_service; then
        log_error "服务启动测试失败"
        exit 1
    fi
    
    log_info "所有测试通过！"
}

# 执行主函数
main 