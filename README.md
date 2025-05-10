# DDNS DNSPod 客户端

## 简介

本项目是一个动态 DNS (DDNS) 客户端，用于自动更新腾讯云 DNSPod 上的 DNS 解析记录。它会定期获取本机的公网 IPv4 和 IPv6 地址，并与 DNSPod 上配置的记录进行比较和更新。

支持作为后台服务运行。

## 特性

- 自动获取公网 IPv4 和 IPv6 地址。
- 支持分别为 IPv4 (A 记录) 和 IPv6 (AAAA 记录) 配置不同的 Record ID 和 SubDomain。
- 通过配置文件 (`config.toml`) 或环境变量进行配置。
- 可作为系统服务运行 (Windows, Linux, macOS)。
- 详细的日志记录，支持日志轮转。

## 配置

应用程序可以通过两种方式进行配置：配置文件和环境变量。环境变量的优先级高于配置文件。

### 1. 配置文件 (`config.toml`)

在应用程序可执行文件所在的目录下创建一个名为 `config.toml` 的文件。如果通过 `-c` 参数指定了其他路径，则会加载指定路径的配置文件。

文件内容示例：

```toml
# 腾讯云 API 密钥
DNSPOD_SECRET_ID = "YOUR_SECRET_ID"
DNSPOD_SECRET_KEY = "YOUR_SECRET_KEY"

# 要操作的主域名
DNSPOD_DOMAIN = "example.com"

# IPv4 (A 记录) 配置
DNSPOD_RECORDID_IPV4 = "123456789"  # A 记录的 Record ID
DNSPOD_SUBDOMAIN_IPV4 = "ddns"     # A 记录的子域名 (例如 ddns.example.com); 如果是主域名本身，请使用 "@"

# IPv6 (AAAA 记录) 配置
DNSPOD_RECORDID_IPV6 = "987654321"  # AAAA 记录的 Record ID
DNSPOD_SUBDOMAIN_IPV6 = "ddns"     # AAAA 记录的子域名 (例如 ddns.example.com); 如果是主域名本身，请使用 "@"
```

**参数说明:**

*   `DNSPOD_SECRET_ID`: 腾讯云账户的 SecretId。
*   `DNSPOD_SECRET_KEY`: 腾讯云账户的 SecretKey。
*   `DNSPOD_DOMAIN`: 您在 DNSPod 上托管的主域名，例如 `example.com`。
*   `DNSPOD_RECORDID_IPV4`: 要更新的 IPv4 (A 记录) 的 Record ID。您可以在 DNSPod 控制台找到它。
*   `DNSPOD_SUBDOMAIN_IPV4`: 与 `DNSPOD_RECORDID_IPV4` 对应的子域名。例如，如果记录是 `www.example.com`，则此处填 `www`。如果是主域名 `@.example.com`，则填 `@`。如果留空，默认为 `@`。
*   `DNSPOD_RECORDID_IPV6`: 要更新的 IPv6 (AAAA 记录) 的 Record ID。
*   `DNSPOD_SUBDOMAIN_IPV6`: 与 `DNSPOD_RECORDID_IPV6` 对应的子域名。如果留空，默认为 `@`。

**注意:**
*   如果某个 IP 类型 (IPv4 或 IPv6) 的 `RECORDID` 未配置或为 `0` (转换后)，则该类型的 DDNS 更新将被跳过。
*   如果 `SUBDOMAIN_IPV4` 或 `SUBDOMAIN_IPV6` 未在配置文件或环境变量中提供，程序会默认使用 `@" `作为对应记录的子域名，代表主域名本身。

### 2. 环境变量

您也可以通过设置以下环境变量来配置应用程序：

*   `DNSPOD_SECRET_ID`
*   `DNSPOD_SECRET_KEY`
*   `DNSPOD_DOMAIN`
*   `DNSPOD_RECORDID_IPV4`
*   `DNSPOD_SUBDOMAIN_IPV4`
*   `DNSPOD_RECORDID_IPV6`
*   `DNSPOD_SUBDOMAIN_IPV6`

## 使用方法

### 直接运行

您可以直接运行编译后的可执行文件：

```bash
./ddns-dnspod
```
或者在 Windows 上：
```cmd
ddns-dnspod.exe
```

如果配置文件不在可执行文件旁边，可以使用 `-c` 参数指定路径：

```bash
./ddns-dnspod -c /path/to/your/config.toml
```

### 作为服务运行

本程序支持作为系统服务运行。

**安装服务:**

```bash
./ddns-dnspod install
```
(Windows 用户请使用管理员权限运行命令提示符或 PowerShell)

**卸载服务:**

```bash
./ddns-dnspod remove
```
(Windows 用户请使用管理员权限运行命令提示符或 PowerShell)

**启动服务:**
(通常由系统服务管理器在安装后或系统启动时自动处理。手动启动方式如下：)

*   Linux (systemd): `sudo systemctl start DDNSDNSPODService`
*   Windows: 在服务管理器 (services.msc) 中找到 "DDNS DNSPOD Service" 并启动，或者使用 `sc start DDNSDNSPODService` (管理员权限)。

**停止服务:**

*   Linux (systemd): `sudo systemctl stop DDNSDNSPODService`
*   Windows: 在服务管理器 (services.msc) 中找到 "DDNS DNSPOD Service" 并停止，或者使用 `sc stop DDNSDNSPODService` (管理员权限)。

**注意:** 服务管理命令通常需要管理员/root权限。服务的具体名称是 `DDNSDNSPODService`。

### 通过 Docker 运行

您也可以通过 Docker 运行此应用程序。推荐使用环境变量来配置 Docker 容器。

**拉取镜像:**
```bash
docker pull ilanyu/ddns-dnspod:latest 
```

**运行容器:**

确保替换以下示例中的占位符为您自己的值。

```bash
docker run -d --name my-ddns-dnspod \
  -e DNSPOD_SECRET_ID="YOUR_SECRET_ID" \
  -e DNSPOD_SECRET_KEY="YOUR_SECRET_KEY" \
  -e DNSPOD_DOMAIN="example.com" \
  -e DNSPOD_RECORDID_IPV4="123456789" \
  -e DNSPOD_SUBDOMAIN_IPV4="ddns" \
  -e DNSPOD_RECORDID_IPV6="987654321" \
  -e DNSPOD_SUBDOMAIN_IPV6="ddns" \
  --restart always \
  ilanyu/ddns-dnspod:latest
```

**必需的环境变量 (在 Docker 中运行时):**

*   `DNSPOD_SECRET_ID`: 您的腾讯云账户 SecretId。
*   `DNSPOD_SECRET_KEY`: 您的腾讯云账户 SecretKey。
*   `DNSPOD_DOMAIN`: 您在 DNSPod 上托管的主域名。
*   `DNSPOD_RECORDID_IPV4`: IPv4 (A 记录) 的 Record ID。如果不需要更新 IPv4，可以省略或留空，但至少需要一个 IPv4 或 IPv6 的 RecordID。
*   `DNSPOD_SUBDOMAIN_IPV4`: (可选) A 记录的子域名。如果省略，默认为 `@` (主域名)。
*   `DNSPOD_RECORDID_IPV6`: IPv6 (AAAA 记录) 的 Record ID。如果不需要更新 IPv6，可以省略或留空。
*   `DNSPOD_SUBDOMAIN_IPV6`: (可选) AAAA 记录的子域名。如果省略，默认为 `@` (主域名)。

**注意:**
*   至少需要配置 `DNSPOD_RECORDID_IPV4` 或 `DNSPOD_RECORDID_IPV6` 中的一个，以便程序执行有效的 DDNS 更新。
*   如果同时使用挂载的 `config.toml` 文件和环境变量，环境变量将覆盖配置文件中的相应值。

**查看日志:**
```bash
docker logs my-ddns-dnspod
```

## 日志

程序运行日志会记录在可执行文件目录下的 `ddns-server.log` 文件中。日志文件会自动轮转，最大大小为 10MB，最多保留 3 个备份，最长保留 7 天。
如果以服务方式运行，日志文件位置可能取决于服务的运行用户和权限设置，但默认仍尝试在可执行文件同级目录创建。
当以交互模式（直接运行）启动时，日志也会同时输出到控制台。

## 构建

如果您拥有 Go 语言开发环境，可以从源码构建：

1.  确保您的项目路径在 `GOPATH` 之外，并且使用了 Go Modules (根目录下有 `go.mod` 文件)。
2.  下载依赖：
    ```bash
    go mod tidy
    ```
3.  构建可执行文件：
    ```bash
    go build -o ddns-dnspod .
    ```
    (在 Windows 上，可以指定输出为 `ddns-dnspod.exe`)

## 贡献

欢迎提交 Pull Requests 或 Issues。

## 许可证

本项目根据 MIT 许可证授权。详情请参阅 `LICENSE` 文件。
