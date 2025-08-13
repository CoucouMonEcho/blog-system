#!/bin/bash

REDIS_VERSION=7.2.4
if command -v redis-server >/dev/null 2>&1 && redis-server --version 2>/dev/null | grep -q "v=${REDIS_VERSION}"; then
  echo "Redis 已是指定版本，跳过下载与编译"
else
  cd /tmp
  wget http://download.redis.io/releases/redis-${REDIS_VERSION}.tar.gz
  tar xzf redis-${REDIS_VERSION}.tar.gz
  cd redis-${REDIS_VERSION}
  make
  sudo make install
fi

# 创建 redis-cluster 配置目录
sudo mkdir -p /opt/redis-cluster
for port in 7001 7002 7003; do
  sudo mkdir -p /opt/redis-cluster/${port}
  cat <<EOF | sudo tee /opt/redis-cluster/${port}/redis.conf
port ${port}
cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 5000
appendonly yes
daemonize yes
bind 0.0.0.0
protected-mode no
logfile /opt/redis-cluster/${port}/redis.log
dir /opt/redis-cluster/${port}
EOF
done

# 启动3个 Redis 实例
for port in 7001 7002 7003; do
  /usr/local/bin/redis-server /opt/redis-cluster/${port}/redis.conf
done

sleep 3

# 创建集群
yes yes | /usr/local/bin/redis-cli --cluster create 127.0.0.1:7001 127.0.0.1:7002 127.0.0.1:7003 --cluster-replicas 0