#!/bin/bash

# Content Service 部署脚本
set -euo pipefail
exec 2>&1

# 配置变量
DEPLOY_PATH="/opt/blog-system"
SERVICE_NAME="content-service"
LOG_PATH="/var/log/blog-system"

log_info() { printf "[INFO] %s\n" "$1"; }
log_error() { printf "[ERROR] %s\n" "$1"; }
silent_exec() { "$@" >/dev/null 2>&1; }

deploy_content_service() {
  log_info "开始部署内容服务..."
  log_info "停止现有服务..."; silent_exec systemctl stop ${SERVICE_NAME} || true

  log_info "更新配置文件..."
  if [ ! -z "${BLOG_PASSWORD:-}" ]; then
    silent_exec sed -i "s/BLOG_PASSWORD/$BLOG_PASSWORD/g" ${DEPLOY_PATH}/configs/content.yaml
    log_info "数据库密码已更新"
  fi
  silent_exec sed -i "s|logs/.*\\.log|${DEPLOY_PATH}/logs/${SERVICE_NAME}.log|g" ${DEPLOY_PATH}/configs/content.yaml
  log_info "日志路径已更新为: ${DEPLOY_PATH}/logs/${SERVICE_NAME}.log"

  log_info "构建应用..."
  cd ${DEPLOY_PATH}/services/content
  export GOOS=linux GOARCH=amd64 CGO_ENABLED=0
  # 统一模块与编译缓存目录
  export GOMODCACHE=${GOMODCACHE:-/opt/blog-system/gomodcache}
  export GOCACHE=${GOCACHE:-/opt/blog-system/gocache}
  mkdir -p "$GOMODCACHE" "$GOCACHE"
  # 仅下载依赖并构建（不执行 tidy）
  silent_exec go mod download
  silent_exec go build -ldflags="-s -w" -o ${SERVICE_NAME} . || go build -o ${SERVICE_NAME} .
  if [ ! -f "${SERVICE_NAME}" ]; then log_error "应用构建失败"; exit 1; fi
  log_info "应用构建成功"

  log_info "创建systemd服务文件..."
  cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Blog System ${SERVICE_NAME}
After=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=${DEPLOY_PATH}/services/content
ExecStart=${DEPLOY_PATH}/services/content/${SERVICE_NAME}
Restart=always
RestartSec=3
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

  silent_exec chmod +x ${DEPLOY_PATH}/services/content/${SERVICE_NAME}
  silent_exec chmod 644 /etc/systemd/system/${SERVICE_NAME}.service
  silent_exec systemctl daemon-reload
  silent_exec systemctl enable ${SERVICE_NAME}
  silent_exec systemctl start ${SERVICE_NAME}
  sleep 8
  if systemctl is-active --quiet ${SERVICE_NAME}; then
    log_info "${SERVICE_NAME} 启动成功"
  else
    log_error "${SERVICE_NAME} 启动失败"; systemctl status ${SERVICE_NAME} --no-pager -l; journalctl -u ${SERVICE_NAME} --no-pager -l; exit 1
  fi
}

main() {
  log_info "开始部署内容服务..."
  if [ "$EUID" -ne 0 ]; then log_error "请使用root权限运行此脚本"; exit 1; fi
  silent_exec mkdir -p ${LOG_PATH} ${DEPLOY_PATH}/logs
  deploy_content_service
  log_info "内容服务部署完成！"
}

main


