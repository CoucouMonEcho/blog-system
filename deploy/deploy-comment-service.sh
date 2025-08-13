#!/bin/bash

# Comment Service 部署脚本
set -euo pipefail
exec 2>&1

DEPLOY_PATH="/opt/blog-system"
SERVICE_NAME="comment-service"
LOG_PATH="/var/log/blog-system"

log_info() { printf "[INFO] %s\n" "$1"; }
log_error() { printf "[ERROR] %s\n" "$1"; }
silent_exec() { "$@" >/dev/null 2>&1; }

deploy_comment_service() {
  log_info "开始部署评论服务..."
  silent_exec systemctl stop ${SERVICE_NAME} || true
  if [ ! -z "${BLOG_PASSWORD:-}" ]; then
    silent_exec sed -i "s/BLOG_PASSWORD/$BLOG_PASSWORD/g" ${DEPLOY_PATH}/configs/comment.yaml
    log_info "数据库密码已更新"
  fi
  silent_exec sed -i "s|logs/.*\\.log|${DEPLOY_PATH}/logs/${SERVICE_NAME}.log|g" ${DEPLOY_PATH}/configs/comment.yaml
  cd ${DEPLOY_PATH}/services/comment
  export GOOS=linux GOARCH=amd64 CGO_ENABLED=0
  silent_exec go mod download
  silent_exec go build -ldflags="-s -w" -o ${SERVICE_NAME} .
  [ -f "${SERVICE_NAME}" ] || { log_error "构建失败"; exit 1; }
  cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Blog System ${SERVICE_NAME}
After=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=${DEPLOY_PATH}/services/comment
ExecStart=${DEPLOY_PATH}/services/comment/${SERVICE_NAME}
Restart=always
RestartSec=3
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
  silent_exec chmod +x ${DEPLOY_PATH}/services/comment/${SERVICE_NAME}
  silent_exec chmod 644 /etc/systemd/system/${SERVICE_NAME}.service
  silent_exec systemctl daemon-reload
  silent_exec systemctl enable ${SERVICE_NAME}
  silent_exec systemctl start ${SERVICE_NAME}
  sleep 8
  systemctl is-active --quiet ${SERVICE_NAME} || { log_error "启动失败"; systemctl status ${SERVICE_NAME} --no-pager -l; exit 1; }
  log_info "${SERVICE_NAME} 启动成功"
}

main() { log_info "开始部署评论服务..."; [ "$EUID" -eq 0 ] || { log_error "需要root权限"; exit 1; }; silent_exec mkdir -p ${LOG_PATH} ${DEPLOY_PATH}/logs; deploy_comment_service; log_info "评论服务部署完成！"; }

main

