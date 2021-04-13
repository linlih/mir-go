# mir-go

## 1. Install

- 要提前装好 minlib，并且minlib与mir-go在同一目录下


```bash
git clone https://gitea.qjm253.cn/PKUSZ-future-network-lab/mir-go.git
cd mir-go
go mod tidy
sudo mkdir /usr/local/etc/mir
sudo cp mirconf.ini /usr/local/etc/mir/mirconf.ini
```

## 2.详细步骤
### 2.1 开启GOMODULE
```bash
go env -w GO111MODULE="on"
```

### 2.2 更新代码
拉取最新代码后，Goland会提示你检测到Go Moudule,点击Enabled即可。
注意：之后的environment可填可不填,需要填的话填写GOPROXY="https://gocenter.io"

### 2.3 修改Go Proxy
```bash
go env -w GOPROXY="https://gocenter.io"
```
再输入
```bash
go env
```
查看状态

### 2.4 更新go mod
```bash
go mod tidy
```

### 2.5 创建本地文件夹
```bash
sudo mkdir /usr/local/etc/mir
```

### 2.6 传入配置文件
```bash
sudo cp mirconf.ini /usr/local/etc/mir/mirconf.ini
```