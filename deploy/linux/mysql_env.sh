#!/bin/bash

if ! rpm -q mysql80-community-release &>/dev/null; then
echo "未检测到 MySQL YUM 源配置包，正在安装..."
sudo rpm -Uvh https://repo.mysql.com/mysql80-community-release-el7-11.noarch.rpm
else
echo "已安装 MySQL YUM 源配置包，跳过下载安装步骤"
fi
sudo rpm --import https://repo.mysql.com/RPM-GPG-KEY-mysql
if ! rpm -q mysql-community-server &>/dev/null; then
sudo yum clean all
sudo yum makecache
fi
if ! rpm -q mysql-community-server &>/dev/null; then
echo "正在安装 MySQL Server..."
sudo yum install -y mysql-server
else
echo "MySQL Server 已安装，跳过"
fi
sudo systemctl enable mysqld
sudo systemctl start mysqld
echo "MySQL root 初始密码如下（如需修改请自行设置）:"
sudo grep 'temporary password' /var/log/mysqld.log || true
echo "MySQL 版本: $(mysql --version)"