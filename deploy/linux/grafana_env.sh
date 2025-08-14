#!/bin/bash

set -euo pipefail

GRAFANA_VERSION=${GRAFANA_VERSION:-11.1.0}
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) GRAF_ARCH=amd64 ;;
  aarch64|arm64) GRAF_ARCH=arm64 ;;
  *) echo "不支持的架构: $ARCH"; exit 1 ;;
esac

TMP_DIR=/tmp
TARBALL=grafana-${GRAFANA_VERSION}.linux-${GRAF_ARCH}.tar.gz
DOWNLOAD_URL=https://dl.grafana.com/oss/release/grafana-${GRAFANA_VERSION}.linux-${GRAF_ARCH}.tar.gz

sudo yum install -y wget tar || sudo apt-get update && sudo apt-get install -y wget tar

cd ${TMP_DIR}
wget --tries=3 --timeout=30 -O ${TARBALL} ${DOWNLOAD_URL}
tar xzf ${TARBALL}

DIR=${TMP_DIR}/grafana-${GRAFANA_VERSION}
sudo mkdir -p /opt/grafana
sudo cp -rf ${DIR}/* /opt/grafana/

sudo tee /etc/systemd/system/grafana.service >/dev/null <<'EOF'
[Unit]
Description=Grafana OSS
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/grafana
ExecStart=/opt/grafana/bin/grafana server --homepath=/opt/grafana
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable grafana
sudo systemctl restart grafana
sleep 2
sudo systemctl status grafana --no-pager || true

echo "Grafana 安装完成，访问 http://localhost:3000 (默认 admin/admin)"


