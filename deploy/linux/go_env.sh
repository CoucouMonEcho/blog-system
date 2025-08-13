#!/bin/bash

GO_VERSION=1.24.0
if command -v go >/dev/null 2>&1 && go version | grep -q "go${GO_VERSION} "; then
  echo "已安装指定 Go 版本，跳过下载和安装"
else
  wget -O /tmp/go${GO_VERSION}.linux-amd64.tar.gz https://mirrors.aliyun.com/golang/go${GO_VERSION}.linux-amd64.tar.gz
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf /tmp/go${GO_VERSION}.linux-amd64.tar.gz
fi
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
export PATH=$PATH:/usr/local/go/bin
go version
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
echo "Go 版本: $(go version)