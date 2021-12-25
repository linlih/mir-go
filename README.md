# mir-go

## 1. Usage

- 首先调用脚本安装

  ```bash
  ./install.sh
  ```

- 然后设置或者修改配置文件中的默认身份 => `/usr/local/etc/mir/mirconf.ini`

- 接着调用 `mirgen` 设置或修改默认身份的密码

  - 验证或者设置：

    ```bash
    sudo mirgen
    ```

  - 修改默认身份的密码：

    ```bash
    sudo mirgen -rp
    ```

  - 如果旧版的密码是明文，可以通过 `-oldPasswdNoHash` 参数兼容

    ```bash
    sudo mirgen -rp -oldPasswdNoHash
    ```

### 1.1 直接启动

```bash
sudo mir
```

### 1.2 安装成系统服务并启动

```bash
# 安装成系统服务
sudo mird install

# 启动程序
sudo mird start

# 终止程序
sudo mird stop

# 查看程序状态 
sudo mird status

# 从系统服务中卸载 => 需要重新覆盖的时候先执行这个
sudo mird remove
```

- 终端日志输出位置
  - Macos
    - `/usr/local/var/log/mird.err`
    - `/usr/local/var/log/mird.log`
  - Linux => `/var/log/mird.log`

关于启动后服务的日志如何输出 => https://blog.csdn.net/sinat_24092079/article/details/120676316

## 2. Install

- 要提前装好 minlib，并且minlib与mir-go在同一目录下


```bash
git clone https://gitea.qjm253.cn/PKUSZ-future-network-lab/mir-go.git
cd mir-go
go mod tidy
sudo mkdir /usr/local/etc/mir
sudo cp mirconf.ini /usr/local/etc/mir/mirconf.ini
```

## 3.详细步骤

### 3.1 开启GOMODULE

```bash
go env -w GO111MODULE="on"
```

### 3.2 更新代码

拉取最新代码后，Goland会提示你检测到Go Moudule,点击Enabled即可。
注意：之后的environment可填可不填,需要填的话填写GOPROXY="https://gocenter.io"

### 3.3 修改Go Proxy

```bash
go env -w GOPROXY="https://gocenter.io"
```
再输入
```bash
go env
```
查看状态

### 3.4 更新go mod

```bash
go mod tidy
```
注意：minlib配置到此操作结束，mir-go配置还需要下面的步骤。

### 3.5 创建本地文件夹

```bash
sudo mkdir /usr/local/etc/mir
```

### 3.6 传入配置文件

```bash
sudo cp mirconf.ini /usr/local/etc/mir/mirconf.ini
```