#!/bin/bash

# Admin Service 部署脚本（仅健康检查）
set -euo pipefail
exec 2>&1

DEPLOY_PATH="/opt/blog-system"
SERVICE_NAME="admin-service"
LOG_PATH="/var/log/blog-system"

log_info() { printf "[INFO] %s\n" "$1"; }
log_error() { printf "[ERROR] %s\n" "$1"; }
silent_exec() { "$@" >/dev/null 2>&1; }

deploy_admin_service() {
  log_info "开始部署管理服务..."
  silent_exec systemctl stop ${SERVICE_NAME} || true
  # 解析端口（从配置提取，默认 8003）
  PORT=$(grep -E '^[[:space:]]*port:' ${DEPLOY_PATH}/configs/admin.yaml 2>/dev/null | head -1 | sed -E 's/.*port:[[:space:]]*([0-9]+).*/\1/')
  PORT=${PORT:-8003}
  # 若端口被占用，强制释放
  {
    ss_bin=$(command -v ss || true)
    netstat_bin=$(command -v netstat || true)
    if [ -n "$ss_bin" ]; then
      pids=$("$ss_bin" -ltnp 2>/dev/null | awk "/:${PORT}\\b/ {print \$0}" | sed -nE 's/.*pid=([0-9]+).*/\1/p' | sort -u)
    elif [ -n "$netstat_bin" ]; then
      pids=$("$netstat_bin" -tlnp 2>/dev/null | awk "/:${PORT}\\b/ {print \$7}" | cut -d'/' -f1 | sort -u)
    else
      pids=""
    fi
    if [ -n "$pids" ]; then
      log_info "端口 ${PORT} 被占用，尝试释放占用进程: $pids"
      for pid in $pids; do
        kill -TERM "$pid" 2>/dev/null || true
      done
      sleep 2
      for pid in $pids; do
        kill -KILL "$pid" 2>/dev/null || true
      done
    fi
  }
  # 更新配置文件（密码与日志路径）
  if [ ! -z "${BLOG_PASSWORD:-}" ]; then
    silent_exec sed -i "s/BLOG_PASSWORD/${BLOG_PASSWORD}/g" ${DEPLOY_PATH}/configs/admin.yaml
    log_info "数据库密码已更新"
  fi
  silent_exec sed -i "s|logs/.*\\.log|${DEPLOY_PATH}/logs/${SERVICE_NAME}.log|g" ${DEPLOY_PATH}/configs/admin.yaml
  log_info "日志路径已更新为: ${DEPLOY_PATH}/logs/${SERVICE_NAME}.log"
  cd ${DEPLOY_PATH}/services/admin
  export GOOS=linux GOARCH=amd64 CGO_ENABLED=0
  # 统一模块与编译缓存目录
  export GOMODCACHE=${GOMODCACHE:-/opt/blog-system/gomodcache}
  export GOCACHE=${GOCACHE:-/opt/blog-system/gocache}
  mkdir -p "$GOMODCACHE" "$GOCACHE"
  # 不 tidy；仅下载依赖并构建
  silent_exec go mod download
  silent_exec go build -ldflags="-s -w" -o ${SERVICE_NAME} . || go build -o ${SERVICE_NAME} .
  [ -f "${SERVICE_NAME}" ] || { log_error "构建失败"; exit 1; }
  cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Blog System ${SERVICE_NAME}
After=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=${DEPLOY_PATH}/services/admin
ExecStart=${DEPLOY_PATH}/services/admin/${SERVICE_NAME}
Restart=always
RestartSec=3
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
  silent_exec chmod +x ${DEPLOY_PATH}/services/admin/${SERVICE_NAME}
  silent_exec chmod 644 /etc/systemd/system/${SERVICE_NAME}.service
  silent_exec systemctl daemon-reload
  silent_exec systemctl enable ${SERVICE_NAME}
  silent_exec systemctl start ${SERVICE_NAME}
  sleep 5
  systemctl is-active --quiet ${SERVICE_NAME} || { \
    log_error "启动失败"; \
    printf "=== 服务状态 ===\n"; systemctl status ${SERVICE_NAME} --no-pager -l; \
    printf "=== 最近日志 ===\n"; journalctl -u ${SERVICE_NAME} -n 200 --no-pager -l || true; \
    printf "=== 配置文件内容 ===\n"; cat ${DEPLOY_PATH}/configs/admin.yaml || true; \
    printf "=== 应用日志(尾部) ===\n"; tail -n 200 ${DEPLOY_PATH}/logs/${SERVICE_NAME}.log || true; \
  printf "=== 端口占用检查 ===\n"; (ss -ltnp 2>/dev/null || netstat -tlnp 2>/dev/null) | grep -E ":${PORT}([^0-9]|$)" || true; \
    printf "=== 可执行文件检查 ===\n"; ls -lah ${DEPLOY_PATH}/services/admin/ || true; file ${DEPLOY_PATH}/services/admin/${SERVICE_NAME} || true; \
    printf "=== 动态链接检查 ===\n"; ldd ${DEPLOY_PATH}/services/admin/${SERVICE_NAME} || true; \
    printf "=== 日志目录 ===\n"; ls -la ${DEPLOY_PATH}/logs/ || true; \
    exit 1; }
  log_info "${SERVICE_NAME} 启动成功"
}

main() { log_info "开始部署管理服务..."; [ "$EUID" -eq 0 ] || { log_error "需要root权限"; exit 1; }; silent_exec mkdir -p ${LOG_PATH} ${DEPLOY_PATH}/logs; deploy_admin_service; log_info "管理服务部署完成！"; }

main

