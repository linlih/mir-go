#!/bin/bash

GOPATH=$(go env GOPATH)

# 一些工具函数
function isMacos() {
  if [ "$(uname)" == "Darwin" ]; then
    return 1
  else
    return 0
  fi
}

function isLinux() {
  if [ "$(uname)" == "Linux" ]; then
    return 1
  else
    return 0
  fi
}

isLinux
linux_platform=$?
isMacos
macos_platform=$?

usr_bin_path=/usr/local/bin

# 创建必要的文件夹
mkdir -p /usr/local/etc/mir
mkdir -p /usr/local/etc/mir/passwd

echo "======================== download ==========================="
if [ $linux_platform -eq 1 ]; then
  # 首先下载依赖
  sudo apt install gcc libpcap-dev -y
  echo ""

elif [ $macos_platform -eq 1 ]; then
  brew install libpcap
fi
go mod download

echo "======================== install mir ==========================="
cp mir "$usr_bin_path"/mir
echo "mir install to $usr_bin_path/mir"
echo ""

echo "======================== install mird ==========================="
cp mird "$usr_bin_path"/mird
echo "mird install to $usr_bin_path/mird"
echo ""

echo "======================== install mirgen ==========================="
cp mirgen "$usr_bin_path"/mirgen
echo "mirgen install to $usr_bin_path/mirgen"
echo ""

echo "======================== compile and install mirc ==========================="
cp mirc "$usr_bin_path"/mirc
echo "mirc install to $usr_bin_path/mirc"
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

if [ $linux_platform -eq 1 ]; then
  echo "======================== copy rysyslog file ==========================="
  # 如果配置文件不存在，则将配置文件拷贝到指定目录下
  if [ ! -f /etc/rsyslog.d/min.conf ]; then
    sudo cp min.conf /etc/rsyslog.d/min.conf
    sudo service rsyslog restart
    echo "rysyslog config file already copy to /etc/rsyslog.d/min.conf"
  else
    echo "rysyslog config file already exists~"
  fi
  echo ""
fi

echo "======================== copy defaultRoute config file ==========================="
# 如果配置文件不存在，则将配置文件拷贝到指定目录下
if [ ! -f /usr/local/etc/mir/defaultRoute.xml ]; then
  sudo cp defaultRoute.xml /usr/local/etc/mir/defaultRoute.xml
  echo "file defaultRoute.xml already copy to /usr/local/etc/mir/defaultRoute.xml"
else
  echo "file defaultRoute.xml already exists~"
fi
echo ""

sudo "$GOPATH"/bin/mirgen
