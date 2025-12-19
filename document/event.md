# 事件系统与启动流程

完整的事件驱动架构说明。

---

## 核心概念

### 1. 事件总线（Event Bus）

基于生产者-消费者模式的事件系统，支持：
- ✅ 异步/同步执行
- ✅ 优先级排序
- ✅ 多消费者订阅
- ✅ 自动注册

### 2. 三个核心包

| 包 | 职责 | 位置 |
|---|------|-----|
| **event** | 事件总线、消费者接口 | `event/` |
| **event_name** | 事件名称常量 | `event_name/` |
| **priority** | 优先级常量 | `priority/` |

---

## 事件系统架构

### Consumer 接口

```go
type Consumer interface {
    // 关注的事件名列表
    CaseEvent() []string
    
    // 事件触发时执行
    Triggered(event *Event) error
    
    // 是否异步执行
    Async() bool
    
    // 优先级（数值越小越先执行）
    Priority() uint32
}
```

### 自动注册机制

**1. 在 consumers 包中通过 init() 注册：**

```go
// config/consumers/consumer.go
func init() {
    event.RegisterConsumer(&ClientCreatedConsumer{})
}
```

**2. 在 app.go 中匿名导入：**

```go
import (
    _ "github.com/charry/config/consumers"  // 自动注册
)
```

**3. event.Init() 时批量注册：**

```go
event.Init(10)  // 自动注册所有消费者
```

---

## 完整的启动流程

### 启动步骤

```
1. LoadEnvArgs()
   ↓
2. config.Init(env)
   ↓
3. event.Init(10)  ← 自动注册所有消费者
   ↓
4. consul.Init(cfg)
   ↓ 发布 ConsulClientCreated 事件
   ├─ [优先级 0] ClientCreatedConsumer
   │   ├─ 从 Consul 加载配置
   │   ├─ 合并配置
   │   └─ 注册 KV 监听
   ├─ [优先级 1] RPCStartConsumer
   │   └─ 启动 RPC 服务器
   └─ [优先级 2] ServiceRegisterConsumer
       └─ 注册服务到 Consul
```

### 启动事件流

```
consul.Init()
    ↓
event.Publish(event_name.ConsulClientCreated, nil)
    ↓ 同步执行，按优先级排序
    ├─ [0] 加载 Consul 配置
    ├─ [1] 启动 RPC 服务器
    └─ [2] 注册服务到 Consul
    ↓
应用启动完成
```

---

## 关闭流程

### 关闭步骤

```
Shutdown()
    ↓
event.Publish(event_name.AppShutdown, nil)
    ↓ 同步执行，按优先级排序
    ├─ [优先级 0] ServiceDeregisterConsumer
    │   └─ 注销服务
    ├─ [优先级 1] RPCStopConsumer
    │   └─ 停止 RPC 服务器
    └─ [优先级 2] ShutdownConsumer
        └─ 停止配置监听
    ↓
event.Close()
    ↓
logger.Sync()
```

---

## 事件列表

### 应用生命周期

| 事件名 | 常量 | 触发时机 | 数据类型 |
|-------|------|---------|---------|
| `app.shutdown` | `event_name.AppShutdown` | 应用关闭 | `nil` |

### Consul 相关

| 事件名 | 常量 | 触发时机 | 数据类型 |
|-------|------|---------|---------|
| `consul.client.created` | `event_name.ConsulClientCreated` | 客户端创建后 | `nil` |
| `consul.kv.changed` | `event_name.ConsulKVChanged` | KV 值变化 | `*consul.KVChangedEvent` |

### 配置相关

| 事件名 | 常量 | 触发时机 | 数据类型 |
|-------|------|---------|---------|
| `config.changed` | `event_name.ConfigChanged` | 配置更新 | `*config.Config` |

---

## 优先级列表

### 启动优先级

| 优先级 | 常量 | 说明 | 消费者 |
|-------|------|------|--------|
| 0 | `priority.ConsulConfigLoad` | 配置加载 | ClientCreatedConsumer |
| 1 | `priority.RPCServerStart` | RPC 启动 | RPCStartConsumer |
| 2 | `priority.ConsulServiceRegister` | 服务注册 | ServiceRegisterConsumer |

### 关闭优先级

| 优先级 | 常量 | 说明 | 消费者 |
|-------|------|------|--------|
| 0 | `priority.ConsulServiceDeregister` | 服务注销 | ServiceDeregisterConsumer |
| 1 | `priority.RPCServerStop` | RPC 停止 | RPCStopConsumer |
| 2 | `priority.ConsulClientClose` | 停止监听 | ShutdownConsumer |

---

## 消费者位置

### 按模块组织

```
config/consumers/
├── ClientCreatedConsumer   # 加载配置
├── KVChangedConsumer       # 监听配置变化
└── ShutdownConsumer        # 停止监听

rpc/consumers/
├── RPCStartConsumer        # 启动 RPC 服务器
└── RPCStopConsumer         # 停止 RPC 服务器

consul/consumers/
├── ServiceRegisterConsumer   # 注册服务
└── ServiceDeregisterConsumer # 注销服务
```

---

## 如何添加新的消费者

### 1. 定义消费者

```go
// mymodule/consumers/consumer.go
package consumers

type MyConsumer struct{}

func (c *MyConsumer) CaseEvent() []string {
    return []string{event_name.ConsulClientCreated}
}

func (c *MyConsumer) Triggered(evt *event.Event) error {
    logger.Info("执行我的初始化逻辑...")
    return nil
}

func (c *MyConsumer) Async() bool {
    return false  // 同步执行
}

func (c *MyConsumer) Priority() uint32 {
    return 10  // 自定义优先级
}

// init 自动注册
func init() {
    event.RegisterConsumer(&MyConsumer{})
}
```

### 2. 在 app.go 中导入

```go
import (
    _ "github.com/charry/mymodule/consumers"  // 自动注册
)
```

### 3. 完成！

无需其他代码，消费者会自动注册并按优先级执行。

---

## 配置热更新流程

### 监听机制

```
1. ClientCreatedConsumer 触发
   ↓
2. consul.RegisterWatch(cfg.AppConfigKey)
   ↓
3. 后台协程监听 Consul KV
   ↓
4. 检测到变化
   ↓
5. event.Publish(event_name.ConsulKVChanged, kvEvent)
   ↓
6. KVChangedConsumer 接收
   ↓
7. 判断 key == AppConfigKey
   ↓
8. config.MergeFromJSON(value)
   ↓
9. event.Publish(event_name.ConfigChanged, cfg)
   ↓
10. 其他消费者接收配置变更通知
```

### 监听任意 KV

```go
// 注册监听
consul.RegisterWatch("myapp/cache-ttl")

// 定义消费者
type CacheTTLConsumer struct{}

func (c *CacheTTLConsumer) CaseEvent() []string {
    return []string{event_name.ConsulKVChanged}
}

func (c *CacheTTLConsumer) Triggered(evt *event.Event) error {
    kvEvt := evt.Data.(*consul.KVChangedEvent)
    if kvEvt.Key == "myapp/cache-ttl" {
        // 更新缓存 TTL
        updateCacheTTL(kvEvt.Value)
    }
    return nil
}
```

---

## 事件驱动的优势

### 1. 模块解耦

```
之前（紧耦合）：
app.go 直接调用各模块的初始化方法

现在（松耦合）：
app.go 只负责发布事件
各模块通过订阅事件响应
```

### 2. 易于扩展

```go
// 添加新功能只需：
// 1. 定义消费者
// 2. 导入包
// 无需修改 app.go
```

### 3. 顺序保证

```go
// 通过优先级精确控制执行顺序
Priority() uint32
```

### 4. 自动注册

```go
// 使用 init() 自动注册
// 无需手动维护注册代码
```

---

## 扩展指南

### 添加新的事件

**1. 在 event_name/event_name.go 中添加：**

```go
const (
    MyNewEvent = "mymodule.new.event"
)
```

**2. 发布事件：**

```go
event.PublishEvent(event_name.MyNewEvent, data)
```

**3. 订阅事件：**

```go
func (c *MyConsumer) CaseEvent() []string {
    return []string{event_name.MyNewEvent}
}
```

### 添加新的优先级

**在 priority/priority.go 中添加：**

```go
const (
    MyModuleInit uint32 = 10  // 自定义优先级
)
```

**使用：**

```go
func (c *MyConsumer) Priority() uint32 {
    return priority.MyModuleInit
}
```

---

## 最佳实践

### 1. 消费者命名

- 启动消费者：`XxxStartConsumer`
- 关闭消费者：`XxxStopConsumer`
- 功能消费者：描述性名称

### 2. 优先级设计

```
0-9: 框架核心初始化
10-99: 基础服务初始化
100+: 业务模块初始化
```

### 3. 事件命名

```
{模块}.{动作}
例如：consul.client.created
     config.changed
     user.registered
```

### 4. 同步 vs 异步

- **同步执行**：关键路径、需要保证顺序
- **异步执行**：日志、统计、非关键操作

---

## 调试技巧

### 查看注册的消费者

```go
logger.Infof("已注册 %d 个消费者", len(pendingConsumers))
```

### 查看事件执行顺序

在消费者的 `Triggered` 方法中添加日志：

```go
func (c *MyConsumer) Triggered(evt *event.Event) error {
    logger.Infof("[优先级 %d] 执行消费者: %T", c.Priority(), c)
    // ...
}
```

---

## 完整示例

### 定义消费者

```go
// mymodule/consumers/consumer.go
package consumers

import (
    "github.com/charry/event"
    "github.com/charry/event_name"
    "github.com/charry/logger"
    "github.com/charry/priority"
)

type DatabaseInitConsumer struct{}

func (c *DatabaseInitConsumer) CaseEvent() []string {
    return []string{event_name.ConsulClientCreated}
}

func (c *DatabaseInitConsumer) Triggered(evt *event.Event) error {
    logger.Info("初始化数据库...")
    // 初始化数据库连接
    return nil
}

func (c *DatabaseInitConsumer) Async() bool {
    return false // 同步执行
}

func (c *DatabaseInitConsumer) Priority() uint32 {
    return 10 // 在 RPC 启动之前
}

func init() {
    event.RegisterConsumer(&DatabaseInitConsumer{})
}
```

### 导入并使用

```go
// app.go
import (
    _ "github.com/charry/mymodule/consumers"  // 自动注册
)

func StartUp() error {
    event.Init(10)    // 自动注册所有消费者
    consul.Init(cfg)  // 触发事件，自动执行
}
```

---

**最后更新**: 2025-12-19

