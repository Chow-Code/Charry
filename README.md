# Charry 项目

微服务开发框架，提供服务发现、配置管理、gRPC 服务器等基础组件。

---

## 🗂️ 模块指引

### 核心模块

| 模块 | 职责 | 文档 |
|------|------|------|
| **config** | 应用配置管理 | 查看代码注释 |
| **logger** | 日志输出 | 查看代码注释 |
| **consul** | 服务注册与发现 | [document/consul.md](document/consul.md) |
| **rpc** | gRPC 服务器封装 | [document/rpc.md](document/rpc.md) |

### 部署文档

| 文档 | 说明 |
|------|------|
| [document/setup.md](document/setup.md) | Consul 服务器部署指南 |

---

## 🚀 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 运行测试程序

#### 方式 1：使用 VS Code 调试（推荐）

按 `F5` 或点击"运行和调试"，选择以下配置之一：

- **Launch Main** - 使用远程 Consul（192.168.30.230:8500）
- **Launch Main (Local Consul)** - 使用本地 Consul（localhost:8500）
- **Launch Main (No Consul)** - 不使用 Consul（测试错误处理）

#### 方式 2：命令行运行

```bash
# 确保 Consul 服务器正在运行
curl http://192.168.30.230:8500/v1/status/leader

# 运行主程序
go run main.go

# 或编译后运行
go build -o charry main.go
./charry
```

### 3. 验证服务注册

```bash
# 在另一个终端查看服务
curl http://192.168.30.230:8500/v1/catalog/service/test-service-dev

# 或访问 Web UI
open http://192.168.30.230:8500/ui
```

### 4. 在代码中使用

```go
package main

import (
    "github.com/charry/config"
    "github.com/charry/rpc"
)

func main() {
    // 创建应用配置
    appConfig := &config.AppConfig{
        Id:          config.LoadIdFromEnv(),      // 从环境变量加载
        Type:        "user-service",               // 代码中设置
        Environment: "prod",                       // 代码中设置
        Addr:        config.LoadAddrFromEnv(),    // 从环境变量加载
    }
    
    // 创建 gRPC 服务器并注册到 Consul
    server, _ := rpc.NewServerWithConsul(nil, appConfig)
    
    // 注册业务服务
    // pb.RegisterUserServiceServer(server.GetGRPCServer(), &userServiceImpl{})
    
    // 启动服务器
    server.StartAsync()
    
    // 等待退出...
    
    // 优雅关闭
    server.Shutdown()
}
```

---

## 📋 开发规范

### 目录结构规范

```
charry/
├── README.md                 # 项目总入口（本文件）
├── main.go                   # 程序入口
├── app.go                    # 应用启动流程
├── go.mod                    # Go 模块定义
├── .gitignore                # Git 忽略文件
├── .vscode/                  # VS Code 配置
│   └── launch.json           # 调试配置（含环境变量）
├── config/                   # 配置模块
│   ├── config.go
│   └── config.example.yaml
├── logger/                   # 日志模块
│   └── logger.go
├── consul/                   # Consul 服务注册模块
│   ├── config.go
│   ├── client.go
│   ├── register.go
│   └── helper.go
├── rpc/                      # RPC 服务器模块
│   ├── server.go
│   ├── options.go
│   └── helper.go
└── document/                 # 📖 统一文档目录
    ├── consul.md             # Consul 模块文档
    ├── rpc.md                # RPC 模块文档
    └── setup.md              # 部署文档
```

### 文档规范

1. **唯一入口**
   - 项目只有一个 README.md（根目录）
   - README.md 职责：模块指引 + 开发规范

2. **模块文档**
   - 每个模块只有一个文档
   - 文档统一放在 `document/` 目录
   - 文件名使用小写字符（如 `consul.md`, `rpc.md`）

3. **文档命名**
   - 模块文档：`{模块名}.md`（如 `consul.md`）
   - 功能文档：`{功能名}.md`（如 `setup.md`）
   - 全小写，使用短横线连接（如需要）

### 模块规范

1. **模块目录**
   - 模块目录下不放 README.md
   - 每个 .go 文件职责单一
   - 文件命名清晰（config.go, client.go, register.go, init.go）

2. **模块初始化**
   - 每个模块都有 `init.go` 文件
   - 提供 `Init()` 方法用于初始化
   - 提供 `Close()` 方法用于清理
   - 使用全局变量（如 GlobalClient, GlobalServer）

3. **模块依赖**
   - 模块之间完全解耦
   - 可选依赖通过独立的文件实现（如 helper.go）

4. **启动流程**
   - 统一在 `app.go` 的 `StartUp()` 方法中管理
   - 按顺序初始化：logger → rpc → consul
   - 按相反顺序关闭：consul → rpc

5. **配置管理**
   - 优先使用环境变量配置
   - 环境变量命名：`模块名_配置项`（如 `CONSUL_ADDRESS`）
   - 提供合理的默认值

### 代码规范

1. **命名规范**
   - 模块名：小写，一个单词（config, consul, rpc）
   - 文件名：小写，下划线分隔（如需要）
   - 类型名：大驼峰（Config, Client, Server）
   - 函数名：大驼峰（公开）或小驼峰（私有）

2. **结构规范**
   - 配置结构放在 config.go
   - 客户端/服务器封装放在对应文件
   - 辅助函数放在 helper.go

3. **错误处理**
   - 使用 `fmt.Errorf` 包装错误
   - 提供清晰的错误信息

### 环境变量规范

```bash
# 通用格式：模块名_配置项（全大写，下划线分隔）

# 应用配置（从环境变量加载）
APP_ID=1                              # 应用实例 ID（默认 1）
APP_HOST=192.168.30.10                # 监听主机（默认 0.0.0.0）
APP_PORT=50051                        # 监听端口（默认 50051）

# Consul 模块
CONSUL_ADDRESS=192.168.30.230:8500    # Consul 地址
CONSUL_DATACENTER=dc1                 # 数据中心（默认 dc1）
CONSUL_HEALTH_CHECK_TYPE=tcp          # 健康检查类型（默认 tcp）
```

**注意**：`Type` 和 `Environment` 在代码中配置，不从环境变量读取。

### 服务命名规范

```
Service ID: {type}-{environment}-{id}
例如: user-service-prod-1

Service Name: {type}-{environment}
例如: user-service-prod

Type 命名: 使用短横线连接（kebab-case）
例如: user-service, order-service, api-gateway
```

---

## 🛠️ 技术栈

- **Go**: 1.25+
- **gRPC**: google.golang.org/grpc
- **Consul**: github.com/hashicorp/consul/api
- **日志**: go.uber.org/zap
- **配置**: github.com/spf13/viper

---

## 📚 文档导航

### 核心模块文档
- [Consul 模块](document/consul.md) - 服务注册与发现
- [RPC 模块](document/rpc.md) - gRPC 服务器封装

### 部署文档
- [Consul 部署指南](document/setup.md) - Consul 服务器安装配置

---

## 🎯 核心特性

- ✅ **完全解耦** - 模块之间零依赖
- ✅ **TCP 健康检查** - 默认使用最简单的方式
- ✅ **环境变量配置** - 12-Factor App 风格
- ✅ **优雅关闭** - 自动处理服务注销
- ✅ **开箱即用** - 极简的 API 设计

---

## 📖 常用命令

### 运行测试程序

```bash
# 运行主程序（测试 Consul 注册）
go run main.go

# 或编译后运行
go build -o charry main.go
./charry
```

### 开发命令

```bash
# 构建项目
go build -v ./...

# 运行测试
go test -v ./...

# 整理依赖
go mod tidy

# 更新依赖
go get -u ./...
```

### 验证服务注册

```bash
# 查看已注册的服务
curl http://192.168.30.230:8500/v1/catalog/services

# 查看服务详情
curl http://192.168.30.230:8500/v1/catalog/service/test-service-dev

# Web UI
open http://192.168.30.230:8500/ui
```

---

## 🤝 贡献规范

1. 遵循上述开发规范
2. 代码提交前确保编译通过
3. 更新相关文档
4. 保持模块解耦

---

**项目维护**: 2025-12-17

