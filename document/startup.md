# 应用启动流程

详细说明 charry 框架的启动和关闭流程。

---

## 启动流程图

```
main()
  ↓
StartUp()
  ↓
1. LoadEnvArgs()
   └─ 加载环境变量（APP_ID, APP_HOST, APP_PORT, etc.）
  ↓
2. config.Init(env)
   ├─ 加载 default.config.json
   └─ 应用环境变量
  ↓
3. event.Init(10)
   ├─ 创建事件总线
   ├─ 启动 10 个工作协程
   └─ 注册所有消费者（自动）
  ↓
4. consul.Init(cfg)
   ├─ 创建 Consul 客户端
   └─ 发布 ConsulClientCreated 事件
       ↓ 同步执行（按优先级）
       ├─ [0] ClientCreatedConsumer
       │   ├─ 从 Consul KV 加载配置
       │   ├─ 合并配置到全局
       │   └─ 注册 KV 监听
       ├─ [1] RPCStartConsumer
       │   └─ 创建并启动 gRPC 服务器
       └─ [2] ServiceRegisterConsumer
           └─ 注册服务到 Consul
  ↓
✓ 应用启动完成
```

---

## 详细步骤说明

### 步骤 1：加载环境变量

```go
env := config.LoadEnvArgs()
```

**加载的环境变量：**
- `APP_ID` - 应用实例 ID
- `APP_HOST` - 监听主机
- `APP_PORT` - 监听端口
- `APP_CONFIG_KEY` - Consul KV 配置键
- `CONSUL_ADDRESS` - Consul 地址
- `CONSUL_DATACENTER` - 数据中心

### 步骤 2：初始化配置

```go
config.Init(env)
```

**配置加载顺序：**
1. 读取 `default.config.json`
2. 应用环境变量覆写
3. 保存到全局配置

### 步骤 3：初始化事件模块

```go
event.Init(10)
```

**自动执行：**
- 创建事件总线
- 启动 10 个工作协程处理异步事件
- 注册所有已通过 init() 预注册的消费者
- 输出：`✓ 已自动注册 N 个事件消费者`

### 步骤 4：初始化 Consul 客户端

```go
consul.Init(cfg)
```

**执行内容：**
1. 创建 Consul API 客户端
2. 测试连接
3. 保存到 `consul.GlobalClient`
4. **发布 `ConsulClientCreated` 事件**

**触发的消费者链（同步执行）：**

#### [优先级 0] ClientCreatedConsumer

```go
// config/consumers/consumer.go
func (c *ClientCreatedConsumer) Triggered(evt *event.Event) error {
    // 1. 从 Consul 加载配置
    jsonStr, _ := consul.GetKV(cfg.AppConfigKey)
    
    // 2. 合并配置
    config.MergeFromJSON(jsonStr)
    
    // 3. 注册 KV 监听
    consul.RegisterWatch(cfg.AppConfigKey)
    
    return nil
}
```

**效果：**
- ✅ 从 Consul 拉取最新配置
- ✅ 合并到本地配置
- ✅ 开始监听配置变化

#### [优先级 1] RPCStartConsumer

```go
// rpc/consumers/consumer.go
func (c *RPCStartConsumer) Triggered(evt *event.Event) error {
    cfg := config.Get()  // 获取合并后的配置
    return rpc.Init(cfg)
}
```

**效果：**
- ✅ 使用最新配置创建 gRPC 服务器
- ✅ 启动服务器

#### [优先级 2] ServiceRegisterConsumer

```go
// consul/consumers/consumer.go
func (c *ServiceRegisterConsumer) Triggered(evt *event.Event) error {
    return consul.Register()
}
```

**效果：**
- ✅ 注册服务到 Consul
- ✅ 注册健康检查

---

## 关闭流程详解

### 触发关闭

```go
Shutdown()
```

### 步骤 1：发布关闭事件

```go
event.PublishEvent(event_name.AppShutdown, nil)
```

**触发的消费者链（同步执行，按优先级）：**

#### [优先级 0] ServiceDeregisterConsumer

```go
func (c *ServiceDeregisterConsumer) Triggered(evt *event.Event) error {
    return consul.Close()
}
```

**效果：**
- ✅ 从 Consul 注销服务
- ✅ 停止 KV 监听

#### [优先级 1] RPCStopConsumer

```go
func (c *RPCStopConsumer) Triggered(evt *event.Event) error {
    rpc.Close()
    return nil
}
```

**效果：**
- ✅ 优雅关闭 gRPC 服务器
- ✅ 等待现有请求完成

#### [优先级 2] ShutdownConsumer

```go
func (c *ShutdownConsumer) Triggered(evt *event.Event) error {
    consul.StopWatch()
    return nil
}
```

**效果：**
- ✅ 停止所有 KV 监听

### 步骤 2：关闭事件模块

```go
event.Close()
```

**效果：**
- 停止所有工作协程
- 关闭事件队列

### 步骤 3：刷新日志

```go
logger.Sync()
```

---

## 配置热更新

### 监听配置变化

**自动监听：**

```
consul.Init()
  ↓
ClientCreatedConsumer
  ↓
consul.RegisterWatch(cfg.AppConfigKey)
  ↓
后台协程监听 KV
```

### 配置变更流程

```
1. Consul KV 值变化
   ↓
2. 后台协程检测到变化
   ↓
3. event.Publish(event_name.ConsulKVChanged, kvEvent)
   ↓
4. KVChangedConsumer 接收
   ↓
5. 判断 key == AppConfigKey
   ↓
6. config.MergeFromJSON(newValue)
   ↓
7. event.Publish(event_name.ConfigChanged, cfg)
   ↓
8. 其他消费者接收并响应配置变化
```

### 订阅配置变更

```go
type MyConfigConsumer struct{}

func (c *MyConfigConsumer) CaseEvent() []string {
    return []string{event_name.ConfigChanged}
}

func (c *MyConfigConsumer) Triggered(evt *event.Event) error {
    cfg := evt.Data.(*config.Config)
    logger.Infof("配置已更新: %s", cfg.App.Type)
    // 处理配置变更...
    return nil
}
```

---

## main.go 代码示例

```go
package main

import (
    _ "github.com/charry/config/consumers"
    _ "github.com/charry/consul/consumers"
    _ "github.com/charry/rpc/consumers"
)

func main() {
    // 启动
    if err := StartUp(); err != nil {
        logger.Fatalf("应用启动失败: %v", err)
    }

    // 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // 关闭
    Shutdown()
}
```

**只需 20 行代码！**

---

## 时序图

### 启动时序

```
Time →
  |
  |-- LoadEnvArgs()
  |-- config.Init()
  |-- event.Init()         [注册所有消费者]
  |-- consul.Init()        [创建客户端]
  |     |
  |     |-- Publish(ConsulClientCreated) [同步]
  |     |     |
  |     |     |-- [0] 加载配置（同步）
  |     |     |-- [1] 启动 RPC（同步）
  |     |     |-- [2] 注册服务（同步）
  |     |
  |     |-- return
  |
  |-- ✓ 启动完成
```

### 关闭时序

```
Time →
  |
  |-- Shutdown()
  |     |
  |     |-- Publish(AppShutdown) [同步]
  |     |     |
  |     |     |-- [0] 注销服务（同步）
  |     |     |-- [1] 停止 RPC（同步）
  |     |     |-- [2] 停止监听（同步）
  |     |
  |     |-- event.Close()
  |     |-- logger.Sync()
  |
  |-- ✓ 关闭完成
```

---

**最后更新**: 2025-12-19

