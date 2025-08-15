#!/bin/bash

set -euo pipefail

# 日志函数
log_info() {
    echo "[INFO] $1"
}

log_error() {
    echo "[ERROR] $1" >&2
}

log_info "开始安装 Prometheus..."

PROM_VERSION=${PROM_VERSION:-2.54.1}
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) PROM_ARCH=amd64 ;;
  aarch64|arm64) PROM_ARCH=arm64 ;;
  *) log_error "不支持的架构: $ARCH"; exit 1 ;;
esac

log_info "检测到架构: $ARCH -> $PROM_ARCH"

TMP_DIR=/tmp
TARBALL=prometheus-${PROM_VERSION}.linux-${PROM_ARCH}.tar.gz
DOWNLOAD_URL=https://ghproxy.net/https://github.com/prometheus/prometheus/releases/download/v${PROM_VERSION}/${TARBALL}

# 检测系统类型并安装依赖
log_info "检查并安装依赖包..."
if command -v yum >/dev/null 2>&1; then
    # RHEL/CentOS/AlmaLinux/Rocky Linux
    log_info "检测到 yum 包管理器，安装依赖..."
    sudo yum install -y wget tar
elif command -v apt-get >/dev/null 2>&1; then
    # Debian/Ubuntu
    log_info "检测到 apt-get 包管理器，安装依赖..."
    sudo apt-get update && sudo apt-get install -y wget tar
elif command -v zypper >/dev/null 2>&1; then
    # openSUSE
    log_info "检测到 zypper 包管理器，安装依赖..."
    sudo zypper install -y wget tar
elif command -v pacman >/dev/null 2>&1; then
    # Arch Linux
    log_info "检测到 pacman 包管理器，安装依赖..."
    sudo pacman -S --noconfirm wget tar
else
    log_error "不支持的包管理器，请手动安装 wget 和 tar"
    exit 1
fi

log_info "下载 Prometheus v${PROM_VERSION}..."
cd ${TMP_DIR}

# 检查是否已经下载过
if [ -f "${TARBALL}" ]; then
    log_info "发现已存在的安装包，跳过下载"
else
    wget --tries=3 --timeout=30 -O ${TARBALL} ${DOWNLOAD_URL}
fi

log_info "解压安装包..."
tar xzf ${TARBALL}

DIR=${TMP_DIR}/prometheus-${PROM_VERSION}.linux-${PROM_ARCH}

log_info "安装 Prometheus 二进制文件..."
sudo cp -f ${DIR}/prometheus /usr/local/bin/
sudo cp -f ${DIR}/promtool /usr/local/bin/
sudo chmod +x /usr/local/bin/prometheus /usr/local/bin/promtool

log_info "创建目录结构..."
sudo mkdir -p /etc/prometheus /var/lib/prometheus
sudo cp -f ${DIR}/consoles/* /etc/prometheus/ 2>/dev/null || true
sudo cp -f ${DIR}/console_libraries/* /etc/prometheus/ 2>/dev/null || true

log_info "创建 Prometheus 配置文件..."
cat >/tmp/prometheus.yml <<'EOF'
global:
  scrape_interval: 5s
  evaluation_interval: 5s

scrape_configs:
  - job_name: 'prometheus'
    static_configs: [ { targets: ['localhost:9090'] } ]
  - job_name: 'blog-gateway'
    static_configs: [ { targets: ['localhost:8000'] } ]
    metrics_path: '/metrics'
  - job_name: 'blog-user'
    static_configs: [ { targets: ['localhost:8001'] } ]
    metrics_path: '/metrics'
  - job_name: 'blog-content'
    static_configs: [ { targets: ['localhost:8002'] } ]
    metrics_path: '/metrics'
  - job_name: 'blog-admin'
    static_configs: [ { targets: ['localhost:8003'] } ]
    metrics_path: '/metrics'
  - job_name: 'blog-stat'
    static_configs: [ { targets: ['localhost:8004'] } ]
    metrics_path: '/metrics'
EOF

sudo mv /tmp/prometheus.yml /etc/prometheus/prometheus.yml

log_info "创建 systemd 服务文件..."
sudo tee /etc/systemd/system/prometheus.service >/dev/null <<'EOF'
[Unit]
Description=Prometheus Monitoring
Documentation=https://prometheus.io/docs/introduction/overview/
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/prometheus \
  --config.file=/etc/prometheus/prometheus.yml \
  --storage.tsdb.path=/var/lib/prometheus \
  --web.console.templates=/etc/prometheus/consoles \
  --web.console.libraries=/etc/prometheus/console_libraries \
  --web.listen-address=:9090 \
  --web.enable-lifecycle
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

log_info "启动 Prometheus 服务..."
sudo systemctl daemon-reload
sudo systemctl enable prometheus
sudo systemctl restart prometheus
sleep 3

log_info "检查服务状态..."
if sudo systemctl is-active --quiet prometheus; then
    log_info "Prometheus 服务启动成功！"
    log_info "访问地址: http://localhost:9090"
    log_info "检查配置: http://localhost:9090/targets"
else
    log_error "Prometheus 服务启动失败"
    sudo systemctl status prometheus --no-pager || true
    exit 1
fi

# 清理临时文件
log_info "清理临时文件..."
rm -f ${TMP_DIR}/${TARBALL}
rm -rf ${DIR}

log_info "Prometheus 安装完成！"


