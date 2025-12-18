# RPC 模块

gRPC 服务器封装模块，提供开箱即用的 gRPC 服务器创建和 Consul 集成。

## 功能特性

- ✅ **简洁高效** - 极简的 gRPC 服务器封装
- ✅ **Consul 集成** - 一行代码完成服务器创建和 Consul 注册
- ✅ **TCP 健康检查** - 默认使用 TCP 健康检查，零配置
- ✅ **优雅关闭** - 自动处理服务注销和服务器停止
- ✅ **灵活配置** - 支持自定义端口、拦截器等选项
- ✅ **完全解耦** - 与 consul 模块完全独立

---

## 快速开始

### 1. 基本使用（不带 Consul）

```go
package main

import (
    "log"
    "github.com/charry/config"
    "github.com/charry/rpc"
)

func main() {
    // 创建应用配置
    appConfig := &config.AppConfig{
        Id:          1,
        Type:        "my-service",
        Environment: "dev",
        Addr: config.Addr{
            Host: "0.0.0.0",
            Port: 50051,
        },
    }

    // 创建 gRPC 服务器（使用默认 RPC 配置）
    server, err := rpc.NewServer(nil, appConfig)
    if err != nil {
        log.Fatal(err)
    }

    // 注册您的业务服务
    // pb.RegisterYourServiceServer(server.GetGRPCServer(), &yourServiceImpl{})

    // 启动服务器
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 2. 与 Consul 集成（推荐）

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/charry/config"
    "github.com/charry/rpc"
)

func main() {
    // 设置环境变量（默认使用 TCP 健康检查）
    os.Setenv("CONSUL_ADDRESS", "192.168.30.230:8500")

    // 创建应用配置
    appConfig := &config.AppConfig{
        Id:          1,
        Type:        "user-service",
        Environment: "prod",
        Addr: config.Addr{
            Host: "192.168.30.10",
            Port: 50051,
        },
    }

    // 创建服务器并注册到 Consul
    // 第一个参数：RPC 配置（nil = 使用默认配置）
    // 第二个参数：应用配置（自动使用其中的 Addr）
    server, err := rpc.NewServerWithConsul(nil, appConfig)
    if err != nil {
        log.Fatal(err)
    }

    // 注册业务服务
    // pb.RegisterUserServiceServer(server.GetGRPCServer(), &userServiceImpl{})

    // 启动服务器
    server.StartAsync()
    log.Println("Server is running...")

    // 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // 优雅关闭（自动从 Consul 注销）
    server.Shutdown()
}
```

---

## API 参考

### 创建服务器

#### `NewServer(rpcConfig *RpcConfig, appConfig *AppConfig) (*Server, error)`

创建一个新的 gRPC 服务器。

**参数：**
- `rpcConfig`: RPC 配置（可传 `nil` 使用默认配置）
- `appConfig`: 应用配置（自动使用其中的 `Addr`）

**示例：**
```go
// 使用默认 RPC 配置
appConfig := &config.AppConfig{
    Addr: config.Addr{Host: "0.0.0.0", Port: 50051},
}
server, err := rpc.NewServer(nil, appConfig)

// 使用自定义 RPC 配置
rpcConfig := &rpc.RpcConfig{
    GrpcOptions: []grpc.ServerOption{
        grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
    },
}
server, err := rpc.NewServer(rpcConfig, appConfig)
```

#### `NewServerWithConsul(rpcConfig *RpcConfig, appConfig *AppConfig) (*ServerWithConsul, error)`

创建 gRPC 服务器并注册到 Consul。

**参数：**
- `rpcConfig`: RPC 配置（可传 `nil` 使用默认配置）
- `appConfig`: 应用配置（自动使用其中的 `Addr`）

**示例：**
```go
// 使用默认配置
server, err := rpc.NewServerWithConsul(nil, appConfig)

// 使用自定义配置
rpcConfig := &rpc.RpcConfig{
    GrpcOptions: []grpc.ServerOption{
        grpc.MaxRecvMsgSize(10 * 1024 * 1024),
    },
}
server, err := rpc.NewServerWithConsul(rpcConfig, appConfig)
```

### 服务器操作

#### `(*Server) Start() error`

启动 gRPC 服务器（阻塞）。

#### `(*Server) StartAsync()`

异步启动 gRPC 服务器（非阻塞）。

#### `(*Server) Stop()`

停止 gRPC 服务器。

#### `(*Server) GetGRPCServer() *grpc.Server`

获取原生 gRPC 服务器，用于注册业务服务。

```go
grpcServer := server.GetGRPCServer()
pb.RegisterUserServiceServer(grpcServer, &userServiceImpl{})
```

### 优雅关闭

#### `(*Server) Shutdown()`

优雅关闭服务器（先标记为不健康，然后停止）。

#### `(*ServerWithConsul) Shutdown()`

优雅关闭服务器并从 Consul 注销。

---

## 使用场景

### 场景 1：简单的 gRPC 服务

```go
func main() {
    appConfig := &config.AppConfig{
        Addr: config.Addr{Host: "0.0.0.0", Port: 50051},
    }
    
    server, _ := rpc.NewServer(nil, appConfig)
    
    // 注册服务
    pb.RegisterGreeterServer(server.GetGRPCServer(), &greeterImpl{})
    
    // 启动
    server.Start()
}
```

### 场景 2：后台服务

```go
func main() {
    appConfig := &config.AppConfig{
        Addr: config.Addr{Host: "0.0.0.0", Port: 50051},
    }
    
    server, _ := rpc.NewServer(nil, appConfig)
    
    // 注册服务
    pb.RegisterUserServiceServer(server.GetGRPCServer(), &userServiceImpl{})
    
    // 异步启动
    server.StartAsync()
    
    // 执行其他初始化...
    initializeOtherComponents()
    
    // 等待退出
    waitForShutdown()
    server.Stop()
}
```

### 场景 3：微服务架构（带 Consul）

```go
func main() {
    appConfig := &config.AppConfig{
        Id:          1,
        Type:        "user-service",
        Environment: "prod",
        Addr: config.Addr{Host: "192.168.30.10", Port: 50051},
    }
    
    // 创建并注册到 Consul
    server, _ := rpc.NewServerWithConsul(nil, appConfig)
    
    // 注册业务服务
    pb.RegisterUserServiceServer(server.GetGRPCServer(), &userServiceImpl{})
    
    // 启动
    server.StartAsync()
    
    // 等待退出
    waitForShutdown()
    
    // 优雅关闭（自动从 Consul 注销）
    server.Shutdown()
}
```

### 场景 4：多服务注册

```go
func main() {
    appConfig := &config.AppConfig{
        Addr: config.Addr{Host: "0.0.0.0", Port: 50051},
    }
    
    server, _ := rpc.NewServer(nil, appConfig)
    
    // 注册多个服务
    pb.RegisterUserServiceServer(server.GetGRPCServer(), &userServiceImpl{})
    pb.RegisterOrderServiceServer(server.GetGRPCServer(), &orderServiceImpl{})
    pb.RegisterPaymentServiceServer(server.GetGRPCServer(), &paymentServiceImpl{})
    
    server.Start()
}
```

---

## 高级配置

### 自定义 gRPC 选项

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    
    "github.com/charry/config"
    "github.com/charry/rpc"
)

func main() {
// TLS 配置
creds, _ := credentials.NewServerTLSFromFile("server.crt", "server.key")

// 拦截器
unaryInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    log.Printf("Calling: %s", info.FullMethod)
    return handler(ctx, req)
}

    // 创建 RPC 配置
    rpcConfig := &rpc.RpcConfig{
        GrpcOptions: []grpc.ServerOption{
            grpc.Creds(creds),                      // TLS
        grpc.UnaryInterceptor(unaryInterceptor), // 拦截器
            grpc.MaxRecvMsgSize(10 * 1024 * 1024),  // 最大接收
            grpc.MaxSendMsgSize(10 * 1024 * 1024),  // 最大发送
        },
    }

    // 创建应用配置
    appConfig := &config.AppConfig{
        Addr: config.Addr{Host: "0.0.0.0", Port: 50051},
    }

    // 创建服务器
    server, _ := rpc.NewServer(rpcConfig, appConfig)
    server.Start()
}
```


---

## 与 Consul 模块的关系

### 模块职责

- **rpc 模块**：负责 gRPC 服务器管理
- **consul 模块**：负责服务注册和发现（默认使用 TCP 健康检查）

### 依赖关系

```
rpc 模块 (可选依赖) → consul 模块
```

两个模块完全解耦：
- `rpc.Server` - 不依赖 consul，可独立使用
- `rpc.ServerWithConsul` - 可选依赖 consul，提供集成功能
- Consul 使用 TCP 健康检查，不依赖 gRPC 健康检查协议

### 使用建议

1. **纯 gRPC 服务**：只使用 `rpc.NewServer()`
2. **微服务架构**：使用 `rpc.NewServerWithConsul()`（默认 TCP 健康检查）

---

## 测试服务

### 使用 grpcurl 测试服务

```bash
# 安装
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 列出服务
grpcurl -plaintext localhost:50051 list

# 调用方法
grpcurl -plaintext -d '{"id": 1}' localhost:50051 myapp.UserService/GetUser
```

---

## 最佳实践

1. **使用 StartAsync() 启动**
   ```go
   server.StartAsync()
   // 继续执行其他初始化...
   ```

2. **处理依赖服务**
   ```go
   // 在服务方法中检查依赖
   func (s *userServiceImpl) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
       if err := db.Ping(); err != nil {
           return nil, status.Error(codes.Unavailable, "database unavailable")
       }
       // 处理请求...
   }
   ```

3. **优雅关闭**
   ```go
   defer server.Shutdown()
   ```

4. **记录服务状态**
   ```go
   func main() {
       server, _ := rpc.NewServerWithConsul(appConfig)
       log.Println("✓ Server registered to Consul with TCP health check")
       
       server.StartAsync()
       log.Println("✓ Server started successfully")
       
       // ... 运行中 ...
       
       server.Shutdown()
       log.Println("✓ Server shutdown complete")
   }
   ```

---

## 故障排查

### 问题：端口已被占用

```bash
# 查看端口占用
lsof -i :50051

# 或使用其他端口
server, _ := rpc.NewServer(rpc.WithPort(50052))
```

### 问题：服务无法连接

```bash
# 测试服务是否启动
telnet localhost 50051

# 或使用 nc
nc -zv localhost 50051
```

---

## 参考资料

- [gRPC Go 文档](https://grpc.io/docs/languages/go/)
- [gRPC 健康检查协议](https://github.com/grpc/grpc/blob/master/doc/health-checking.md)
- [Consul 模块文档](./consul.md)
- [项目 README](../README.md)

---

**最后更新**: 2025-12-17

