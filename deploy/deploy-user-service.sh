#!/bin/bash

# User Service 部署脚本
set -euo pipefail
exec 2>&1

# 配置变量
DEPLOY_PATH="/opt/blog-system"
SERVICE_NAME="user-service"
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

# 部署用户服务
deploy_user_service() {
	log_info "开始部署用户服务..."
	
	# 停止现有服务
	log_info "停止现有服务..."
	silent_exec systemctl stop ${SERVICE_NAME} || true
	
	# 更新配置文件
	log_info "更新配置文件..."
	if [ ! -z "$BLOG_PASSWORD" ]; then
		silent_exec sed -i "s/BLOG_PASSWORD/$BLOG_PASSWORD/g" ${DEPLOY_PATH}/configs/user.yaml
		log_info "数据库密码已更新"
	fi
	
	# 修复日志路径
	silent_exec sed -i "s|logs/.*\.log|${DEPLOY_PATH}/logs/${SERVICE_NAME}.log|g" ${DEPLOY_PATH}/configs/user.yaml
	log_info "日志路径已更新为: ${DEPLOY_PATH}/logs/${SERVICE_NAME}.log"
	
	# 安装二进制
	log_info "安装二进制..."
	cd ${DEPLOY_PATH}/services/user
	# 如果 CI 已上传二进制则直接使用，否则回退到构建
	if [ ! -f "${SERVICE_NAME}" ]; then
		export GOOS=linux
		export GOARCH=amd64
		export CGO_ENABLED=0
		export GOMODCACHE=${GOMODCACHE:-/opt/blog-system/gomodcache}
		export GOCACHE=${GOCACHE:-/opt/blog-system/gocache}
		mkdir -p "$GOMODCACHE" "$GOCACHE"
		log_info "未检测到二进制，回退为远端构建"
		silent_exec go mod download
		silent_exec go build -ldflags="-s -w" -o ${SERVICE_NAME} . || go build -o ${SERVICE_NAME} .
	fi
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
		cat ${DEPLOY_PATH}/configs/user.yaml
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

    # HTTP 健康检查工具
    local curl_bin=""
    if command -v curl >/dev/null 2>&1; then
        curl_bin=$(command -v curl)
    else
        for p in /usr/bin/curl /bin/curl; do
            [ -x "$p" ] && curl_bin="$p" && break
        done
    fi
    local health_url="http://127.0.0.1:${port}/health"

    for i in $(seq 1 ${retries}); do
        # 1) 如果有 curl，优先做 HTTP 就绪检查
        if [ -n "$curl_bin" ]; then
            if "$curl_bin" -fsS -m 2 "$health_url" 2>/dev/null | grep -q '"status"'; then
                log_info "${service_name} 健康检查通过 (${health_url})"
                return 0
            fi
        fi
        # 2) 回退端口层面检查
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

# 主函数
main() {
	log_info "开始部署用户服务..."
	
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
	
	# 部署用户服务
	deploy_user_service
	
	# 检查用户服务端口
	check_port 8001 "用户服务"
	
	log_info "用户服务部署完成！"
	log_info "用户服务地址: http://$(hostname -I | awk '{print $1}'):8001"
	log_info "日志文件: ${DEPLOY_PATH}/logs/"
	printf "\n"
}

# 执行主函数
main 