#!/bin/bash

# Stat Service 部署脚本
set -euo pipefail
exec 2>&1

DEPLOY_PATH="/opt/blog-system"
SERVICE_NAME="stat-service"
LOG_PATH="/var/log/blog-system"

log_info() { printf "[INFO] %s\n" "$1"; }
log_error() { printf "[ERROR] %s\n" "$1"; }
silent_exec() { "$@" >/dev/null 2>&1; }

deploy_stat_service() {
  log_info "开始部署统计服务..."
  silent_exec systemctl stop ${SERVICE_NAME} || true
  # 释放端口占用（从配置解析，默认 8004）
  PORT=$(grep -E '^[[:space:]]*port:' ${DEPLOY_PATH}/configs/stat.yaml 2>/dev/null | head -1 | sed -E 's/.*port:[[:space:]]*([0-9]+).*/\1/')
  PORT=${PORT:-8004}
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
    log_info "端口 ${PORT} 被占用，释放占用 PID: $pids"
    for pid in $pids; do kill -TERM "$pid" 2>/dev/null || true; done
    sleep 2
    for pid in $pids; do kill -KILL "$pid" 2>/dev/null || true; done
  fi
  silent_exec sed -i "s|logs/.*\\.log|${DEPLOY_PATH}/logs/${SERVICE_NAME}.log|g" ${DEPLOY_PATH}/configs/stat.yaml
  cd ${DEPLOY_PATH}/services/stat
  # 如果 CI 已上传二进制则直接使用，否则回退到构建
  if [ ! -f "${SERVICE_NAME}" ]; then
    export GOOS=linux GOARCH=amd64 CGO_ENABLED=0
    export GOMODCACHE=${GOMODCACHE:-/opt/blog-system/gomodcache}
    export GOCACHE=${GOCACHE:-/opt/blog-system/gocache}
    mkdir -p "$GOMODCACHE" "$GOCACHE"
    log_info "未检测到二进制，回退为远端构建"
    # 依赖整理与按需下载
    MOD_BEFORE=$(sha256sum go.mod 2>/dev/null | awk '{print $1}')
    SUM_BEFORE=$(sha256sum go.sum 2>/dev/null | awk '{print $1}')
    silent_exec go mod tidy
    MOD_AFTER=$(sha256sum go.mod 2>/dev/null | awk '{print $1}')
    SUM_AFTER=$(sha256sum go.sum 2>/dev/null | awk '{print $1}')
    DOWNLOADED=0
    if [ "${MOD_BEFORE}" != "${MOD_AFTER}" ] || [ "${SUM_BEFORE}" != "${SUM_AFTER}" ]; then
      log_info "检测到依赖变更，下载依赖..."
      silent_exec go mod download
      DOWNLOADED=1
    fi
    # 构建，失败则补充下载后重试
    if ! silent_exec go build -ldflags="-s -w" -o ${SERVICE_NAME} .; then
      if [ "$DOWNLOADED" -eq 0 ]; then
        log_info "构建失败，下载依赖后重试..."
        silent_exec go mod download
      fi
      go build -ldflags="-s -w" -o ${SERVICE_NAME} .
    fi
  fi
  [ -f "${SERVICE_NAME}" ] || { log_error "构建失败"; exit 1; }
  cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Blog System ${SERVICE_NAME}
After=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=${DEPLOY_PATH}/services/stat
ExecStart=${DEPLOY_PATH}/services/stat/${SERVICE_NAME}
Restart=always
RestartSec=3
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
  silent_exec chmod +x ${DEPLOY_PATH}/services/stat/${SERVICE_NAME}
  silent_exec chmod 644 /etc/systemd/system/${SERVICE_NAME}.service
  silent_exec systemctl daemon-reload
  silent_exec systemctl enable ${SERVICE_NAME}
  silent_exec systemctl start ${SERVICE_NAME}
  sleep 8
  systemctl is-active --quiet ${SERVICE_NAME} || { log_error "启动失败"; systemctl status ${SERVICE_NAME} --no-pager -l; exit 1; }
  log_info "${SERVICE_NAME} 启动成功"
}

main() { log_info "开始部署统计服务..."; [ "$EUID" -eq 0 ] || { log_error "需要root权限"; exit 1; }; silent_exec mkdir -p ${LOG_PATH} ${DEPLOY_PATH}/logs; deploy_stat_service; log_info "统计服务部署完成！"; }

main

