# Charry 事件系统

一个高性能、易于使用的Go语言事件订阅和发布系统。

## 功能特性

- 🚀 **高性能**: 使用worker池处理事件，支持并发处理
- 🎯 **灵活订阅**: 支持事件类型过滤和自定义过滤器
- 🔧 **可扩展处理**: 基于函数的处理器设计，支持自定义逻辑
- 🔗 **链式处理**: 支持同步和异步链式处理器
- 📊 **统计监控**: 提供详细的统计信息和监控数据
- ⚡ **同步/异步**: 支持同步和异步事件发布
- 🛡️ **错误处理**: 完善的错误处理和日志记录

## 快速开始

### 1. 创建事件管理器

```go
import "charry/event"

// 创建事件管理器，使用3个worker处理事件
eventManager := event.NewEventManager(3)

// 启动事件管理器
if err := eventManager.Start(); err != nil {
    log.Fatal("启动事件管理器失败:", err)
}
defer eventManager.Stop()
```

### 2. 创建和订阅事件

```go
// 创建函数处理器
logHandler := event.NewFunctionHandler(
    "日志处理器",
    func(ctx context.Context, event event.Event) error {
        log.Printf("处理事件: %s, 数据: %v", event.Type, event.Data)
        return nil
    },
    func(eventType string) bool {
        return true // 处理所有事件类型
    },
)

// 订阅用户注册事件
subscriptionId, err := eventManager.Subscribe("user.registered", logHandler)
if err != nil {
    log.Printf("订阅失败: %v", err)
}
```

### 3. 发布事件

```go
// 创建事件数据
userData := map[string]interface{}{
    "user_id": "user_001",
    "username": "张三",
    "email": "zhangsan@example.com",
}

// 创建事件
userEvent := event.NewEvent("user.registered", "user-service", userData).
    WithMetadata("ip", "192.168.1.100").
    WithMetadata("user_agent", "Chrome/91.0")

// 异步发布事件
if err := eventManager.Publish(userEvent); err != nil {
    log.Printf("发布事件失败: %v", err)
}

// 或者同步发布事件
ctx := context.Background()
if err := eventManager.PublishSync(ctx, userEvent); err != nil {
    log.Printf("同步发布事件失败: %v", err)
}
```

## 处理器类型

### 函数处理器

函数处理器是最灵活的处理器类型，允许你使用自定义函数处理事件：

```go
// 创建邮件通知处理器
emailHandler := event.NewFunctionHandler(
    "邮件通知处理器",
    func(ctx context.Context, event event.Event) error {
        // 实现邮件发送逻辑
        log.Printf("发送邮件通知: %s", event.Type)
        return sendEmail(event.Data)
    },
    func(eventType string) bool {
        return eventType == "user.registered" || eventType == "order.completed"
    },
)

// 创建数据库保存处理器
dbHandler := event.NewFunctionHandler(
    "数据库保存处理器",
    func(ctx context.Context, event event.Event) error {
        // 实现数据库保存逻辑
        return saveToDatabase(event)
    },
    func(eventType string) bool {
        return strings.HasPrefix(eventType, "user.") || strings.HasPrefix(eventType, "order.")
    },
)

// 创建HTTP webhook处理器
webhookHandler := event.NewFunctionHandler(
    "Webhook处理器",
    func(ctx context.Context, event event.Event) error {
        // 实现HTTP请求逻辑
        payload, _ := json.Marshal(event)
        return sendWebhook("https://api.example.com/webhook", payload)
    },
    func(eventType string) bool {
        return eventType == "payment.completed"
    },
)
```

### 链式处理器

链式处理器按顺序执行多个处理器：

```go
// 同步链式处理器 - 按顺序执行
chainHandler := event.NewChainHandler(false, // 不在错误时停止
    emailHandler,
    dbHandler,
    webhookHandler,
)

// 异步链式处理器 - 并发执行
asyncChainHandler := event.NewAsyncChainHandler(5*time.Second,
    emailHandler,
    dbHandler,
    webhookHandler,
)
```

## 高级功能

### 事件过滤器
```go
// 创建过滤器，只处理高优先级事件
priorityFilter := func(e event.Event) bool {
    if priority, exists := e.Metadata["priority"]; exists {
        return priority == "high" || priority == "critical"
    }
    return false
}

// 使用过滤器订阅
eventManager.Subscribe("system.error", errorHandler, priorityFilter)
```

### 统计信息
```go
// 获取统计信息
stats := eventManager.GetStats()
fmt.Printf("统计信息: %+v\n", stats)

// 获取订阅信息
subscriptions := eventManager.GetSubscriptions()
for eventType, subs := range subscriptions {
    fmt.Printf("事件类型 %s 有 %d 个订阅者\n", eventType, len(subs))
}
```

## 事件结构

```go
type Event struct {
    Id        string                 `json:"id"`        // 事件唯一标识
    Type      string                 `json:"type"`      // 事件类型
    Data      interface{}            `json:"data"`      // 事件数据
    Timestamp time.Time              `json:"timestamp"` // 事件时间戳
    Source    string                 `json:"source"`    // 事件源
    Metadata  map[string]interface{} `json:"metadata"`  // 元数据
}
```

## 实际应用示例

### 用户注册流程
```go
// 创建多个处理器处理用户注册
func setupUserRegistration(em *event.EventManager) {
    // 发送欢迎邮件
    welcomeHandler := event.NewFunctionHandler(
        "欢迎邮件处理器",
        func(ctx context.Context, event event.Event) error {
            userData := event.Data.(UserData)
            return sendWelcomeEmail(userData.Email, userData.Username)
        },
        func(eventType string) bool {
            return eventType == "user.registered"
        },
    )

    // 创建用户目录
    directoryHandler := event.NewFunctionHandler(
        "用户目录处理器",
        func(ctx context.Context, event event.Event) error {
            userData := event.Data.(UserData)
            return createUserDirectory(userData.UserId)
        },
        func(eventType string) bool {
            return eventType == "user.registered"
        },
    )

    // 更新统计信息
    statsHandler := event.NewFunctionHandler(
        "统计更新处理器",
        func(ctx context.Context, event event.Event) error {
            return updateUserStats()
        },
        func(eventType string) bool {
            return eventType == "user.registered"
        },
    )

    // 使用异步链式处理器并发处理
    registrationChain := event.NewAsyncChainHandler(10*time.Second,
        welcomeHandler,
        directoryHandler,
        statsHandler,
    )

    em.Subscribe("user.registered", registrationChain)
}
```

### 订单处理流程
```go
func setupOrderProcessing(em *event.EventManager) {
    // 订单处理链
    orderChain := event.NewChainHandler(true, // 遇到错误时停止
        event.NewFunctionHandler("库存检查", checkInventory, orderEventFilter),
        event.NewFunctionHandler("价格计算", calculatePrice, orderEventFilter),
        event.NewFunctionHandler("订单保存", saveOrder, orderEventFilter),
        event.NewFunctionHandler("发送确认", sendOrderConfirmation, orderEventFilter),
    )

    em.Subscribe("order.created", orderChain)
}

func orderEventFilter(eventType string) bool {
    return strings.HasPrefix(eventType, "order.")
}
```

## 运行示例

```bash
# 运行完整示例
go run example_main.go

# 运行测试
go test ./event -v

# 检查覆盖率
go test ./event -cover
```

## 最佳实践

1. **事件命名**: 使用有意义的事件名称，建议使用点分格式，如 `user.registered`, `order.created`
2. **错误处理**: 总是检查订阅和发布的错误返回值
3. **资源管理**: 确保在程序退出前调用 `Stop()` 方法
4. **性能优化**: 根据业务需求调整worker池大小
5. **处理器设计**: 保持处理器逻辑简单，避免长时间运行的操作
6. **监控**: 定期检查统计信息，监控系统状态

## 注意事项

- 事件管理器使用worker池处理异步事件，确保在高并发场景下有足够的worker数量
- 同步发布会阻塞当前协程，建议在需要立即处理结果的场景下使用
- 函数处理器应该避免长时间运行，以免影响整体性能
- 建议为关键业务逻辑设置超时时间和错误重试机制
- 使用链式处理器时要注意错误处理策略的选择

## 贡献

欢迎提交Issue和Pull Request来改进这个项目！