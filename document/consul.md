# Consul 模块使用指南

## 模块概述

本项目新增了 `consul` 模块，用于将应用配置（AppConfig）自动注册到 Consul 服务发现系统。该模块支持服务注册、健康检查、服务发现和优雅关闭等功能。

## 核心功能

- ✅ **自动服务注册** - 将服务信息注册到 Consul
- ✅ **健康检查** - 自动配置 HTTP 健康检查
- ✅ **服务发现** - 查询和发现其他服务实例
- ✅ **环境变量配置** - 通过环境变量配置 Consul 地址
- ✅ **优雅关闭** - 退出时自动注销服务
- ✅ **元数据支持** - 支持自定义服务元数据

---

## 快速开始

### 1. 设置环境变量

```bash
# 应用配置（可选，有默认值）
export APP_ID="1"                     # 默认 1
export APP_HOST="192.168.30.10"       # 默认 0.0.0.0
export APP_PORT="50051"               # 默认 50051

# Consul 配置
export CONSUL_ADDRESS="192.168.30.230:8500"

# 可选配置（有默认值）
export CONSUL_DATACENTER="dc1"
export CONSUL_HEALTH_CHECK_TYPE="tcp"
```

### 2. 在代码中使用

```go
package main

import (
    "log"
    "os/signal"
    "syscall"
    
    "github.com/charry/config"
    "github.com/charry/consul"
)

func main() {
    // 1. 创建应用配置
    appConfig := &config.AppConfig{
        Id:          config.LoadIdFromEnv(),      // 从环境变量加载
        Type:        "api-server",                 // 代码中设置
        Environment: "prod",                       // 代码中设置
        Addr:        config.LoadAddrFromEnv(),    // 从环境变量加载
        Metadata: map[string]any{
            "version": "1.0.0",
        },
    }
    
    // 2. 注册服务到 Consul
    client, err := consul.RegisterFromEnv(appConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. 优雅关闭时注销服务
    defer client.GracefulShutdown(appConfig)
    
    // 4. 应用主逻辑...
    log.Println("Service running...")
    
    // 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
}
```

### 3. 实现健康检查端点

模块会自动配置 HTTP 健康检查，您需要在应用中实现 `/health` 端点：

```go
import "net/http"

http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    // 检查应用健康状态（数据库连接、依赖服务等）
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})

http.ListenAndServe(":8080", nil)
```

---

## 模块结构

```
consul/
├── config.go         # Consul 配置（从环境变量加载）
├── client.go         # Consul 客户端封装
├── register.go       # 服务注册、注销、查询逻辑
├── helper.go         # 便捷方法
└── README.md         # 模块详细文档
```

---

## 环境变量说明

| 环境变量 | 必需 | 默认值 | 说明 |
|---------|------|--------|------|
| `CONSUL_ADDRESS` | ✅ | `localhost:8500` | Consul 服务器地址 |
| `CONSUL_DATACENTER` | ❌ | `dc1` | 数据中心名称 |
| `CONSUL_HEALTH_CHECK_INTERVAL` | ❌ | `10s` | 健康检查间隔 |
| `CONSUL_HEALTH_CHECK_TIMEOUT` | ❌ | `5s` | 健康检查超时 |
| `CONSUL_DEREGISTER_CRITICAL_SERVICE_AFTER` | ❌ | `30s` | 不健康服务注销时间 |

---

## 服务注册规则

### Service ID（服务唯一标识）
格式：`{type}-{environment}-{id}`

示例：
- `api-server-prod-1`
- `worker-test-2`

### Service Name（服务名称）
格式：`{type}-{environment}`

同类服务共享相同名称，便于服务发现。

示例：
- `api-server-prod` - 生产环境所有 API 服务器
- `worker-test` - 测试环境所有 Worker

### 健康检查
- **URL**: `http://{host}:{port}/health`
- **方法**: HTTP GET
- **间隔**: 10 秒（可配置）
- **超时**: 5 秒（可配置）

### 服务标签（Tags）
自动生成的标签：
- `id:{id}` - 实例 ID
- `type:{type}` - 服务类型
- `env:{environment}` - 环境
- `{key}:{value}` - Metadata 中的键值对

---

## API 参考

### 便捷方法

#### `RegisterFromEnv(appConfig *AppConfig) (*Client, error)`
从环境变量创建客户端并注册服务（推荐使用）。

```go
client, err := consul.RegisterFromEnv(appConfig)
```

#### `MustRegisterFromEnv(appConfig *AppConfig) *Client`
同上，但失败时会 panic。

#### `(*Client) GracefulShutdown(appConfig *AppConfig)`
优雅关闭时注销服务。

```go
defer client.GracefulShutdown(appConfig)
```

### 客户端方法

#### `NewClient(cfg *Config) (*Client, error)`
使用自定义配置创建客户端。

#### `(*Client) Ping() error`
测试 Consul 连接。

### 服务注册方法

#### `(*Client) RegisterService(appConfig *AppConfig) error`
注册服务。

#### `(*Client) DeregisterService(appConfig *AppConfig) error`
注销服务。

### 服务发现方法

#### `(*Client) GetHealthyService(serviceName string) ([]*ServiceEntry, error)`
获取健康的服务实例。

```go
services, err := client.GetHealthyService("api-server-prod")
for _, s := range services {
    fmt.Printf("%s:%d\n", s.Service.Address, s.Service.Port)
}
```

#### `(*Client) ListServices() (map[string][]string, error)`
列出所有已注册服务。

---

## 使用场景

### 场景 1：微服务注册

```go
// 在微服务启动时注册
func main() {
    appConfig := loadConfig()
    client, _ := consul.RegisterFromEnv(appConfig)
    defer client.GracefulShutdown(appConfig)
    
    startServer()
}
```

### 场景 2：服务发现

```go
// 查找其他服务实例
func callUserService() {
    services, err := client.GetHealthyService("user-service-prod")
    if err != nil || len(services) == 0 {
        return err
    }
    
    // 使用第一个健康实例
    service := services[0]
    url := fmt.Sprintf("http://%s:%d/api/user",
        service.Service.Address,
        service.Service.Port)
    
    // 调用服务...
}
```

### 场景 3：多实例部署

```go
// 实例 1
appConfig1 := &config.AppConfig{
    Id: 1,
    Type: "api-server",
    Environment: "prod",
    Addr: config.Addr{Host: "192.168.30.10", Port: 8080},
}

// 实例 2
appConfig2 := &config.AppConfig{
    Id: 2,
    Type: "api-server",
    Environment: "prod",
    Addr: config.Addr{Host: "192.168.30.11", Port: 8080},
}

// 两个实例都注册到 "api-server-prod"，客户端可负载均衡
```

---

## 部署配置

### Docker 部署

```yaml
# docker-compose.yml
version: '3.8'
services:
  myapp:
    image: myapp:latest
    environment:
      - CONSUL_ADDRESS=192.168.30.230:8500
      - CONSUL_DATACENTER=dc1
    ports:
      - "8080:8080"
```

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: myapp
        image: myapp:latest
        env:
        - name: CONSUL_ADDRESS
          value: "consul.default.svc.cluster.local:8500"
        ports:
        - containerPort: 8080
```

### Systemd 服务

```ini
# /etc/systemd/system/myapp.service
[Unit]
Description=My Application
After=network.target

[Service]
Type=simple
User=myapp
Environment="CONSUL_ADDRESS=192.168.30.230:8500"
Environment="CONSUL_DATACENTER=dc1"
ExecStart=/usr/local/bin/myapp
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

---

## 在 Consul Web UI 中查看

1. 访问：http://192.168.30.230:8500/ui
2. 点击左侧 **Services** 菜单
3. 找到您的服务（如 `api-server-prod`）
4. 查看：
   - 服务实例列表
   - 健康状态
   - 标签和元数据
   - 健康检查详情

---

## 验证服务注册

```bash
# 设置环境变量
export CONSUL_ADDRESS="192.168.30.230:8500"

# 查看所有已注册的服务
curl http://192.168.30.230:8500/v1/catalog/services

# 查看特定服务详情
curl http://192.168.30.230:8500/v1/catalog/service/your-service-name
```

---

## 最佳实践

### 1. 优雅关闭
始终使用 `defer` 确保服务注销：
```go
defer client.GracefulShutdown(appConfig)
```

### 2. 健康检查
实现有意义的健康检查，不仅返回 200：
```go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    // 检查数据库连接
    if err := db.Ping(); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        return
    }
    
    // 检查依赖服务
    if !isDependencyHealthy() {
        w.WriteHeader(http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})
```

### 3. 元数据
使用元数据传递版本、区域等信息：
```go
Metadata: map[string]any{
    "version": "1.2.3",
    "region": "cn-east",
    "team": "backend",
    "git_commit": "abc123",
}
```

### 4. 服务命名
使用清晰的服务类型和环境：
- Type: `api-server`, `worker`, `scheduler`
- Environment: `dev`, `test`, `staging`, `prod`

### 5. 错误处理
妥善处理注册失败：
```go
client, err := consul.RegisterFromEnv(appConfig)
if err != nil {
    log.Printf("Failed to register to Consul: %v", err)
    // 决定是否继续运行（某些场景可以降级运行）
}
```

---

## 故障排查

### 服务注册失败

**问题**: `failed to register service`

**排查步骤**:
```bash
# 1. 检查 Consul 连接
curl http://192.168.30.230:8500/v1/status/leader

# 2. 检查环境变量
echo $CONSUL_ADDRESS

# 3. 检查网络连接
ping 192.168.30.230
telnet 192.168.30.230 8500
```

### 健康检查失败

**问题**: 服务显示为 Critical

**排查步骤**:
```bash
# 1. 测试健康检查端点
curl http://your-service:8080/health

# 2. 查看 Consul 日志
docker logs consul-server

# 3. 检查服务健康检查配置
curl http://192.168.30.230:8500/v1/agent/checks
```

### 服务未出现在 Consul

**排查步骤**:
```bash
# 1. 列出所有服务
curl http://192.168.30.230:8500/v1/agent/services

# 2. 查看应用日志
# 确认是否有 "Service registered successfully" 日志

# 3. 检查服务 ID 是否冲突
# Service ID 必须唯一
```

---

## 相关文档

- [RPC 模块文档](./rpc.md) - gRPC 服务器封装
- [Consul 部署指南](./setup.md) - Consul 服务器部署
- [项目 README](../README.md) - 项目总入口

---

## 技术支持

如遇到问题，请查看：
1. 模块 README 文档
2. 示例程序代码
3. Consul 官方文档：https://www.consul.io/docs

---

**文档更新时间**: 2025-12-16

