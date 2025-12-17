# Consul 服务发现与配置管理部署指南

## 服务器信息

- **服务器地址**: 192.168.30.230
- **操作系统**: CentOS Linux 7
- **Docker 版本**: 20.10.24
- **Consul 版本**: 1.22.1

---

## 一、代理配置

### 1.1 Docker 守护进程代理（已配置）

Docker 已配置为通过代理拉取镜像，无需手动干预。

**配置文件**: `/etc/systemd/system/docker.service.d/http-proxy.conf`

```ini
[Service]
Environment="HTTP_PROXY=http://127.0.0.1:7890"
Environment="HTTPS_PROXY=http://127.0.0.1:7890"
Environment="NO_PROXY=localhost,127.0.0.1,::1,192.168.0.0/16,10.0.0.0/8"
```

### 1.2 终端代理命令（已配置）

服务器已在 `~/.bashrc` 中配置以下命令：

| 命令 | 功能 |
|------|------|
| `proxy` | 开启终端代理（HTTP/HTTPS/SOCKS） |
| `unproxy` | 关闭终端代理 |
| `proxystat` | 查看当前代理状态 |
| `docker-proxy-status` | 查看 Docker 代理配置 |
| `docker-proxy-reload` | 重启 Docker 服务（应用代理配置） |

**使用示例**:

```bash
# 开启代理后下载文件
proxy
curl https://www.google.com
git clone https://github.com/user/repo.git
unproxy

# 查看代理状态
proxystat
```

---

## 二、Consul 部署

### 2.1 部署信息

- **部署方式**: Docker Compose
- **配置文件**: `/root/consul.docker.compose.yml`
- **数据目录**: `/root/consul/data`
- **配置目录**: `/root/consul/config`
- **容器名称**: consul-server

### 2.2 Docker Compose 配置

```yaml
version: '3.8'

services:
  consul:
    image: hashicorp/consul:latest
    container_name: consul-server
    restart: unless-stopped
    ports:
      - "8500:8500"  # HTTP API 和 Web UI
      - "8600:8600/udp"  # DNS
      - "8300:8300"  # Server RPC
    volumes:
      - ./consul/data:/consul/data
      - ./consul/config:/consul/config
    command: agent -server -ui -bootstrap-expect=1 -client=0.0.0.0
    environment:
      - CONSUL_BIND_INTERFACE=eth0
```

### 2.3 服务管理命令

```bash
# 进入工作目录
cd /root

# 启动 Consul
docker-compose -f consul.docker.compose.yml up -d

# 停止 Consul
docker-compose -f consul.docker.compose.yml down

# 重启 Consul
docker-compose -f consul.docker.compose.yml restart

# 查看日志
docker-compose -f consul.docker.compose.yml logs -f

# 查看容器状态
docker ps | grep consul
docker logs consul-server
```

---

## 三、访问 Consul

### 3.1 Web UI（推荐）

浏览器访问：**http://192.168.30.230:8500**

功能包括：
- 服务注册和发现管理
- 健康检查监控
- Key/Value 配置管理（可视化编辑）
- 节点状态查看
- ACL 权限管理

### 3.2 HTTP API

#### 基础操作

```bash
# 获取集群状态
curl http://192.168.30.230:8500/v1/status/leader

# 获取集群成员
curl http://192.168.30.230:8500/v1/agent/members

# 获取所有服务
curl http://192.168.30.230:8500/v1/catalog/services
```

#### Key/Value 配置管理

```bash
# 写入配置（单个值）
curl -X PUT -d 'value' http://192.168.30.230:8500/v1/kv/myapp/config/key

# 读取配置（原始值）
curl http://192.168.30.230:8500/v1/kv/myapp/config/key?raw

# 读取配置（JSON 格式，包含元数据）
curl http://192.168.30.230:8500/v1/kv/myapp/config/key

# 删除配置
curl -X DELETE http://192.168.30.230:8500/v1/kv/myapp/config/key

# 批量读取（递归）
curl http://192.168.30.230:8500/v1/kv/myapp?recurse
```

#### 从文件导入配置

```bash
# 从 JSON 文件导入
curl -X PUT --data-binary @config.json \
  http://192.168.30.230:8500/v1/kv/myapp/config
```

---

## 四、配置迁移示例

### 4.1 原配置结构（config.json）

```json
{
  "prod": {
    "1.1.0": {
      "audit": false,
      "global": "https://g.com"
    },
    "1.2.0": "https://g.com",
    "default": "https://g.com"
  },
  "test": {
    "default": "https://g.com"
  }
}
```

### 4.2 迁移到 Consul（方式一：通过 API）

```bash
# 生产环境配置
curl -X PUT -d 'false' \
  http://192.168.30.230:8500/v1/kv/myapp/prod/1.1.0/audit

curl -X PUT -d 'https://g.com' \
  http://192.168.30.230:8500/v1/kv/myapp/prod/1.1.0/global

curl -X PUT -d 'https://g.com' \
  http://192.168.30.230:8500/v1/kv/myapp/prod/1.2.0

curl -X PUT -d 'https://g.com' \
  http://192.168.30.230:8500/v1/kv/myapp/prod/default

# 测试环境配置
curl -X PUT -d 'https://g.com' \
  http://192.168.30.230:8500/v1/kv/myapp/test/default
```

### 4.3 迁移到 Consul（方式二：通过 Web UI）

1. 访问 http://192.168.30.230:8500
2. 点击左侧菜单 **Key/Value**
3. 点击 **Create** 按钮
4. 输入 Key 和 Value，例如：
   - Key: `myapp/prod/1.1.0/audit`
   - Value: `false`
5. 点击 **Save** 保存

### 4.4 读取配置示例

```bash
# 读取单个配置
curl http://192.168.30.230:8500/v1/kv/myapp/prod/default?raw
# 输出: https://g.com

# 读取整个环境配置
curl http://192.168.30.230:8500/v1/kv/myapp/prod?recurse

# 在应用中使用（示例：Python）
import requests

response = requests.get('http://192.168.30.230:8500/v1/kv/myapp/prod/default?raw')
config_value = response.text
print(config_value)
```

---

## 五、服务注册与发现

### 5.1 注册服务（HTTP API）

```bash
# 注册一个服务
curl -X PUT -d '{
  "ID": "myapp-01",
  "Name": "myapp",
  "Tags": ["v1.0", "production"],
  "Address": "192.168.30.10",
  "Port": 8080,
  "Check": {
    "HTTP": "http://192.168.30.10:8080/health",
    "Interval": "10s"
  }
}' http://192.168.30.230:8500/v1/agent/service/register
```

### 5.2 查询服务

```bash
# 查询所有服务
curl http://192.168.30.230:8500/v1/catalog/services

# 查询特定服务的实例
curl http://192.168.30.230:8500/v1/catalog/service/myapp

# 只查询健康的服务实例
curl http://192.168.30.230:8500/v1/health/service/myapp?passing
```

### 5.3 注销服务

```bash
curl -X PUT http://192.168.30.230:8500/v1/agent/service/deregister/myapp-01
```

---

## 六、健康检查

### 6.1 支持的健康检查类型

- **HTTP**: 定期发送 HTTP 请求检查
- **TCP**: 检查 TCP 端口是否可连接
- **Script**: 执行自定义脚本
- **TTL**: 服务自行报告健康状态
- **gRPC**: gRPC 健康检查协议

### 6.2 健康检查示例

```bash
# HTTP 检查
curl -X PUT -d '{
  "Name": "myapp-health",
  "ServiceID": "myapp-01",
  "HTTP": "http://192.168.30.10:8080/health",
  "Interval": "10s",
  "Timeout": "5s"
}' http://192.168.30.230:8500/v1/agent/check/register

# TCP 检查
curl -X PUT -d '{
  "Name": "mydb-health",
  "TCP": "192.168.30.20:3306",
  "Interval": "10s"
}' http://192.168.30.230:8500/v1/agent/check/register
```

---

## 七、常见问题

### 7.1 无法访问 Web UI

**问题**: 浏览器无法访问 http://192.168.30.230:8500

**解决方案**:
```bash
# 1. 检查容器是否运行
docker ps | grep consul

# 2. 检查端口映射
netstat -tlnp | grep 8500

# 3. 检查防火墙
firewall-cmd --list-ports
firewall-cmd --permanent --add-port=8500/tcp
firewall-cmd --reload

# 4. 查看容器日志
docker logs consul-server
```

### 7.2 Docker 无法拉取镜像

**问题**: 提示网络错误或超时

**解决方案**:
```bash
# 检查代理配置
docker-proxy-status

# 重启 Docker 服务
docker-proxy-reload

# 测试代理
proxy
curl -I https://hub.docker.com
```

### 7.3 配置未生效

**解决方案**:
```bash
# 重启 Consul 容器
cd /root
docker-compose -f consul.docker.compose.yml restart

# 或重新部署
docker-compose -f consul.docker.compose.yml down
docker-compose -f consul.docker.compose.yml up -d
```

---

## 八、Consul 主要功能

### 8.1 服务发现
- 服务注册与注销
- 服务健康检查
- DNS 和 HTTP 接口查询
- 多数据中心支持

### 8.2 配置管理（Key/Value Store）
- 动态配置存储
- 版本控制
- 分布式锁
- 功能开关（Feature Flags）

### 8.3 健康检查
- HTTP/TCP/gRPC 检查
- 脚本检查
- TTL 检查
- 自动故障转移

### 8.4 服务网格（Consul Connect）
- 服务间加密通信（mTLS）
- 访问控制（Intentions）
- 流量管理
- 可观测性

---

## 九、最佳实践

### 9.1 配置管理
- 使用层级化的 Key 结构（如 `app/env/version/key`）
- 敏感信息使用 Consul 的 ACL 保护
- 定期备份配置数据

### 9.2 服务注册
- 为服务添加有意义的标签（Tags）
- 配置合适的健康检查间隔
- 使用服务 ID 区分同一服务的不同实例

### 9.3 运维监控
- 定期查看 Consul 日志
- 监控服务健康状态
- 设置数据备份计划

---

## 十、参考资源

- **官方文档**: https://www.consul.io/docs
- **HTTP API 文档**: https://www.consul.io/api-docs
- **配置文件位置**: `/root/consul.docker.compose.yml`
- **服务器代理配置**: `~/.bashrc`
- **Docker 代理配置**: `/etc/systemd/system/docker.service.d/http-proxy.conf`

---

## 快速命令速查

```bash
# 启动 Consul
cd /root && docker-compose -f consul.docker.compose.yml up -d

# 查看状态
docker ps | grep consul

# 查看日志
docker logs -f consul-server

# 访问 Web UI
# http://192.168.30.230:8500

# 写入配置
curl -X PUT -d 'value' http://192.168.30.230:8500/v1/kv/key

# 读取配置
curl http://192.168.30.230:8500/v1/kv/key?raw

# 开启终端代理
proxy

# 查看代理状态
proxystat
```

---

**文档最后更新时间**: 2025-12-16

