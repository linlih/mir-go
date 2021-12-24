#!/bin/bash

GOPATH=$(go env GOPATH)

mkdir -p /usr/local/etc/mir
mkdir -p /usr/local/etc/mir/passwd

# 首先下载依赖
echo "======================== download ==========================="
sudo apt install gcc libpcap-dev -y
go mod download
echo ""

echo "======================== compile and install mir ==========================="
go install ./daemon/mircmd/mir
echo "mir install to $GOPATH/bin/mir"
echo ""

echo "======================== compile and install mird ==========================="
go install ./daemon/mircmd/mird
echo "mir install to $GOPATH/bin/mird"
echo ""

echo "======================== compile and install mirgen ==========================="
go install ./daemon/mircmd/mirgen
echo "mirc install to $GOPATH/bin/mirgen"
echo ""

echo "======================== compile and install mirc ==========================="
go install ./daemon/mgmt/mirc
echo "mirc install to $GOPATH/bin/mirc"
echo ""

echo "======================== copy config file ==========================="
# 如果配置文件不存在，则将配置文件拷贝到指定目录下
if [ ! -f /usr/local/etc/mir/mirconf.ini ]; then
    sudo cp mirconf.ini /usr/local/etc/mir/mirconf.ini
    echo "config file already copy to /usr/local/etc/mir/mirconf.ini"
else
    echo "config file already exists~"
fi
echo ""
echo "======================== copy defaultRoute config file ==========================="
# 如果配置文件不存在，则将配置文件拷贝到指定目录下
if [ ! -f /usr/local/etc/mir/defaultRoute.xml ]; then
    sudo cp defaultRoute.xml /usr/local/etc/mir/defaultRoute.xml
    echo "file defaultRoute.xml already copy to /usr/local/etc/mir/defaultRoute.xml"
else
    echo "file defaultRoute.xml already exists~"
fi
echo ""

sudo $GOPATH/bin/mirgen