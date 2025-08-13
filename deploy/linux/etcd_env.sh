#!/bin/bash

ETCD_VERSION=${ETCD_VERSION:-3.5.14}
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ETCD_ARCH=amd64 ;;
  aarch64|arm64) ETCD_ARCH=arm64 ;;
  *) echo "不支持的架构: $ARCH"; exit 1 ;;
esac

TMP_DIR=/tmp
TARBALL=etcd-v${ETCD_VERSION}-linux-${ETCD_ARCH}.tar.gz
DOWNLOAD_URL=https://ghproxy.net/https://github.com/etcd-io/etcd/releases/download/v${ETCD_VERSION}/${TARBALL}

sudo yum install -y wget tar

cd ${TMP_DIR}
if ! wget --tries=3 --timeout=30 -O ${TARBALL} ${DOWNLOAD_URL}; then
  echo "下载 etcd 失败: ${DOWNLOAD_URL}" >&2
  exit 1
fi
if ! tar xzf ${TARBALL}; then
  echo "解压失败: ${TARBALL}" >&2
  exit 1
fi

sudo cp -f ${TMP_DIR}/etcd-v${ETCD_VERSION}-linux-${ETCD_ARCH}/etcd /usr/local/bin/
sudo cp -f ${TMP_DIR}/etcd-v${ETCD_VERSION}-linux-${ETCD_ARCH}/etcdctl /usr/local/bin/
sudo chmod +x /usr/local/bin/etcd /usr/local/bin/etcdctl

if ! id -u etcd >/dev/null 2>&1; then
  sudo useradd --system --home-dir /var/lib/etcd --shell /sbin/nologin etcd
fi
sudo mkdir -p /var/lib/etcd
sudo chown -R etcd:etcd /var/lib/etcd

sudo tee /etc/systemd/system/etcd.service >/dev/null <<'EOF'
[Unit]
Description=etcd key-value store
Documentation=https://etcd.io
After=network.target

[Service]
Type=notify
User=etcd
ExecStart=/usr/local/bin/etcd \
  --name etcd \
  --data-dir /var/lib/etcd \
  --initial-advertise-peer-urls http://127.0.0.1:2380 \
  --listen-peer-urls http://0.0.0.0:2380 \
  --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://127.0.0.1:2379 \
  --initial-cluster etcd=http://127.0.0.1:2380 \
  --initial-cluster-state new \
  --initial-cluster-token etcd-cluster-1
Restart=on-failure
RestartSec=5
LimitNOFILE=40000

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable etcd
sudo systemctl restart etcd

sleep 2
sudo systemctl status etcd --no-pager || true

export ETCDCTL_API=3
etcdctl version || true
etcdctl --endpoints=http://127.0.0.1:2379 endpoint health || true

echo "etcd 安装与启动完成。" 