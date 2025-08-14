#!/bin/bash

set -euo pipefail

PROM_VERSION=${PROM_VERSION:-2.54.1}
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) PROM_ARCH=amd64 ;;
  aarch64|arm64) PROM_ARCH=arm64 ;;
  *) echo "不支持的架构: $ARCH"; exit 1 ;;
esac

TMP_DIR=/tmp
TARBALL=prometheus-${PROM_VERSION}.linux-${PROM_ARCH}.tar.gz
DOWNLOAD_URL=https://ghproxy.net/https://github.com/prometheus/prometheus/releases/download/v${PROM_VERSION}/${TARBALL}

sudo yum install -y wget tar || sudo apt-get update && sudo apt-get install -y wget tar

cd ${TMP_DIR}
wget --tries=3 --timeout=30 -O ${TARBALL} ${DOWNLOAD_URL}
tar xzf ${TARBALL}

DIR=${TMP_DIR}/prometheus-${PROM_VERSION}.linux-${PROM_ARCH}
sudo cp -f ${DIR}/prometheus /usr/local/bin/
sudo cp -f ${DIR}/promtool /usr/local/bin/
sudo chmod +x /usr/local/bin/prometheus /usr/local/bin/promtool

sudo mkdir -p /etc/prometheus /var/lib/prometheus
sudo cp -f ${DIR}/consoles/* /etc/prometheus/ 2>/dev/null || true
sudo cp -f ${DIR}/console_libraries/* /etc/prometheus/ 2>/dev/null || true

cat >/tmp/prometheus.yml <<'EOF'
global:
  scrape_interval: 5s

scrape_configs:
  - job_name: 'blog-gateway'
    static_configs: [ { targets: ['localhost:8000'] } ]
  - job_name: 'blog-user'
    static_configs: [ { targets: ['localhost:8001'] } ]
  - job_name: 'blog-content'
    static_configs: [ { targets: ['localhost:8002'] } ]
  - job_name: 'blog-admin'
    static_configs: [ { targets: ['localhost:8003'] } ]
  - job_name: 'blog-stat'
    static_configs: [ { targets: ['localhost:8004'] } ]
EOF

sudo mv /tmp/prometheus.yml /etc/prometheus/prometheus.yml

sudo tee /etc/systemd/system/prometheus.service >/dev/null <<'EOF'
[Unit]
Description=Prometheus Monitoring
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/prometheus \
  --config.file=/etc/prometheus/prometheus.yml \
  --storage.tsdb.path=/var/lib/prometheus \
  --web.listen-address=:9090
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable prometheus
sudo systemctl restart prometheus
sleep 2
sudo systemctl status prometheus --no-pager || true

echo "Prometheus 安装完成，访问 http://localhost:9090"


