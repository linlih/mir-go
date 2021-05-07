#!/bin/bash

# 首先下载依赖
echo "======================== download ==========================="
go mod download
echo ""

echo "======================== compile and install mir ==========================="
go install ./daemon/mir
echo "mir install to $GOPATH/bin/mir"
echo ""

echo "======================== compile and install mirc ==========================="
go install ./daemon/mgmt/mirc
echo "mirc install to $GOPATH/bin/mirc"
echo ""

echo "======================== copy config file ==========================="
# 如果配置文件不存在，则将配置文件拷贝到指定目录下
if [ ! -f /usr/local/etc/mir/mirconf.ini ]; then
    sudo mkdir -p /usr/local/etc/mir
    sudo cp mirconf.ini /usr/local/etc/mir/mirconf.ini
    echo "config file already copy to /usr/local/etc/mir/mirconf.ini"
else
    echo "config file already exists~"
fi
echo ""
