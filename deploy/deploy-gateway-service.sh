#!/bin/bash

# Gateway Service 部署脚本
set -euo pipefail
exec 2>&1

# 配置变量
DEPLOY_PATH="/opt/blog-system"
SERVICE_NAME="gateway-service"
LOG_PATH="/var/log/blog-system"

# 日志函数
log_info() {
    printf "[INFO] %s\n" "$1"
}

log_error() {
    printf "[ERROR] %s\n" "$1"
}

# 静默执行函数 - 完全抑制输出
silent_exec() {
    "$@" >/dev/null 2>&1
}

# 静默执行函数 - 只抑制标准输出
silent_exec_stdout() {
    "$@" >/dev/null
}

# 部署网关服务
deploy_gateway_service() {
    log_info "开始部署网关服务..."
    
    # 停止现有服务
    log_info "停止现有服务..."
    silent_exec systemctl stop ${SERVICE_NAME} || true
    
    # 更新配置文件
    log_info "更新配置文件..."
    if [ ! -z "$BLOG_PASSWORD" ]; then
        silent_exec sed -i "s/BLOG_PASSWORD/$BLOG_PASSWORD/g" ${DEPLOY_PATH}/configs/gateway.yaml
        log_info "数据库密码已更新"
    fi
    
    # 修复日志路径
    silent_exec sed -i "s|logs/.*\.log|${DEPLOY_PATH}/logs/${SERVICE_NAME}.log|g" ${DEPLOY_PATH}/configs/gateway.yaml
    log_info "日志路径已更新为: ${DEPLOY_PATH}/logs/${SERVICE_NAME}.log"
    
    # 构建应用
    log_info "构建应用..."
    cd ${DEPLOY_PATH}/services/gateway
    
    export GOOS=linux
    export GOARCH=amd64
    export CGO_ENABLED=0
    
    # 静默执行go命令
    silent_exec go mod download
    silent_exec go build -ldflags="-s -w" -o ${SERVICE_NAME} .
    
    if [ ! -f "${SERVICE_NAME}" ]; then
        log_error "应用构建失败"
        exit 1
    fi
    log_info "应用构建成功"
    
    # 创建systemd服务文件
    log_info "创建systemd服务文件..."
    
    cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Blog System ${SERVICE_NAME}
After=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=${DEPLOY_PATH}/services/gateway
ExecStart=${DEPLOY_PATH}/services/gateway/${SERVICE_NAME}
Restart=always
RestartSec=3
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    # 设置权限
    silent_exec chmod +x ${DEPLOY_PATH}/services/gateway/${SERVICE_NAME}
    silent_exec chmod 644 /etc/systemd/system/${SERVICE_NAME}.service
    
    # 重新加载systemd并启动服务
    silent_exec systemctl daemon-reload
    silent_exec systemctl enable ${SERVICE_NAME}
    silent_exec systemctl start ${SERVICE_NAME}
    
    # 等待服务启动
    sleep 8
    
    # 检查服务状态
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        log_info "${SERVICE_NAME} 启动成功"
    else
        log_error "${SERVICE_NAME} 启动失败"
        printf "=== 服务状态 ===\n"
        systemctl status ${SERVICE_NAME} --no-pager -l
        printf "=== 服务日志 ===\n"
        journalctl -u ${SERVICE_NAME} --no-pager -l
        printf "=== 配置文件内容 ===\n"
        cat ${DEPLOY_PATH}/configs/gateway.yaml
        printf "=== 日志文件 ===\n"
        ls -la ${DEPLOY_PATH}/logs/ || printf "日志目录不存在\n"
        exit 1
    fi
}

# 检查端口函数
check_port() {
    local port=$1
    local service_name=$2
    local retries=${CHECK_PORT_RETRIES:-30}
    local interval=${CHECK_PORT_INTERVAL:-2}

    # 解析可执行路径，避免 PATH 缺少 /usr/sbin 时找不到 ss/netstat
    local ss_bin=""
    if command -v ss >/dev/null 2>&1; then
        ss_bin=$(command -v ss)
    else
        for p in /usr/sbin/ss /usr/bin/ss /sbin/ss; do
            [ -x "$p" ] && ss_bin="$p" && break
        done
    fi
    local netstat_bin=""
    if command -v netstat >/dev/null 2>&1; then
        netstat_bin=$(command -v netstat)
    else
        for p in /usr/sbin/netstat /usr/bin/netstat /bin/netstat; do
            [ -x "$p" ] && netstat_bin="$p" && break
        done
    fi

    for i in $(seq 1 ${retries}); do
        if [ -n "$ss_bin" ]; then
            if "$ss_bin" -ltn 2>/dev/null | grep -Eq ":${port}([^0-9]|$)"; then
                log_info "${service_name} 端口${port}监听正常"
                return 0
            fi
        elif [ -n "$netstat_bin" ]; then
            if "$netstat_bin" -tlnp 2>/dev/null | grep -Eq ":${port}([^0-9]|$)"; then
                log_info "${service_name} 端口${port}监听正常"
                return 0
            fi
        fi
        sleep ${interval}
    done
    log_error "${service_name} 端口${port}未监听"
    return 1
}

# 检查依赖服务
check_dependencies() {
    log_info "检查依赖服务..."
    
    # 检查用户服务是否运行
    if ! systemctl is-active --quiet user-service; then
        log_error "用户服务未运行，请先部署用户服务"
        exit 1
    fi
    
    # 检查用户服务端口
    if ! silent_exec netstat -tlnp | silent_exec grep -q ":8001"; then
        log_error "用户服务端口8001未监听，请检查用户服务"
        exit 1
    fi
    
    log_info "依赖服务检查通过"
}

# 主函数
main() {
    log_info "开始部署网关服务..."
    
    # 检查是否为root用户
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用root权限运行此脚本"
        exit 1
    fi
    
    # 创建目录结构
    log_info "创建目录结构..."
    silent_exec mkdir -p ${LOG_PATH}
    silent_exec mkdir -p ${DEPLOY_PATH}/logs
    
    # 检查Redis Cluster配置
    log_info "检查Redis Cluster配置..."
    if ! command -v redis-cli &> /dev/null; then
        log_error "Redis客户端未安装，请先安装Redis"
        exit 1
    fi
    
    # 检查Redis Cluster节点
    for port in 7001 7002 7003; do
        if silent_exec redis-cli -p $port ping; then
            log_info "Redis Cluster节点 $port 正常"
        else
            log_error "Redis Cluster节点 $port 不可用"
            log_info "请确保Redis Cluster已正确配置并运行在端口 7001, 7002, 7003"
        fi
    done
    
    # 检查依赖服务
    check_dependencies
    
    # 部署网关服务
    deploy_gateway_service
    
    # 检查网关服务端口
    check_port 8000 "网关服务"
    
    log_info "网关服务部署完成！"
    log_info "网关服务地址: http://$(hostname -I | awk '{print $1}'):8000"
    log_info "日志文件: ${DEPLOY_PATH}/logs/"
    printf "\n"
}

# 执行主函数
main 