# Config 模块

配置管理模块，提供环境变量加载和配置合并功能。

---

## 核心功能

### 1. EnvArgs - 统一环境变量管理

`EnvArgs` 是应用启动后第一个加载的结构体，所有环境变量统一从这里获取。

**结构定义：**
```go
type EnvArgs struct {
    // 应用配置
    AppId   uint16
    AppType string  // 应用类型
    AppHost string
    AppPort int

    // Consul 配置
    ConsulAddress             string
    ConsulDatacenter          string
    ConsulHealthCheckType     string
    // ... 其他 Consul 配置
}
```

**使用方法：**
```go
// 启动后第一步：加载环境变量
env := config.LoadEnvArgs()

// 第二步：创建完整配置
cfg := config.NewConfigFromEnv(env)

// cfg 包含：
// - cfg.App (应用配置)
// - cfg.Consul (Consul 配置)
```

### 2. MergeConfig - 配置合并

支持将两个 Config 对象合并，config2 中不为空的值会覆盖 config1。

**使用场景：**
- 从 Consul 读取配置并覆盖本地配置
- 多环境配置合并
- 动态配置更新

**示例：**
```go
// 本地配置
localConfig := &config.Config{
    App: config.AppConfig{
        Id:          1,
        Type:        "user-service",
        Environment: "prod",
        Addr: config.Addr{
            Host: "0.0.0.0",
            Port: 50051,
        },
    },
}

// 从 Consul 读取的配置（只有部分字段）
consulConfig := &config.Config{
    App: config.AppConfig{
        Addr: config.Addr{
            Port: 8080,  // 只覆盖端口
        },
        Metadata: map[string]any{
            "region": "cn-east",
        },
    },
}

// 合并配置
finalConfig := config.MergeConfig(localConfig, consulConfig)

// 结果：
// - Id: 1 (来自 localConfig)
// - Type: "user-service" (来自 localConfig)
// - Environment: "prod" (来自 localConfig)
// - Host: "0.0.0.0" (来自 localConfig)
// - Port: 8080 (被 consulConfig 覆盖)
// - Metadata: {"region": "cn-east"} (来自 consulConfig)
```

---

## API 参考

### 环境变量加载

#### `LoadEnvArgs() *EnvArgs`

加载所有环境变量到 EnvArgs 结构体。

**支持的环境变量：**
- `APP_ID` - 应用实例 ID（默认 1）
- `APP_TYPE` - 应用类型（默认 app-service）
- `APP_HOST` - 监听主机（默认 0.0.0.0）
- `APP_PORT` - 监听端口（默认 50051）
- `CONSUL_ADDRESS` - Consul 地址（默认 localhost:8500）
- `CONSUL_DATACENTER` - 数据中心（默认 dc1）
- `CONSUL_HEALTH_CHECK_TYPE` - 健康检查类型（默认 tcp）
- 其他 Consul 相关配置...

#### `LoadIdFromEnv(env *EnvArgs) uint16`

从 EnvArgs 加载应用 ID。

#### `LoadAddrFromEnv(env *EnvArgs) Addr`

从 EnvArgs 加载监听地址。

### 配置合并

#### `MergeConfig(config1, config2 *Config) *Config`

合并两个 Config 对象。

**合并规则：**
- config2 中不为零值的字段会覆盖 config1
- 零值字段保持 config1 的值
- Metadata 会合并（不是替换）
- 直接修改 config1 并返回其引用

---

## 使用流程

### 标准启动流程

```go
func main() {
    // 1. 加载环境变量（第一步）
    env := config.LoadEnvArgs()
    
    // 2. 从环境变量创建完整配置
    cfg := config.NewConfigFromEnv(env)
    
    // 3. 设置应用特定配置
    cfg.App.Type = "user-service"
    cfg.App.Environment = "prod"
    
    // 4. 启动应用（只传 Config）
    StartUp(cfg)
}
```

### 从 Consul 读取并合并配置

```go
// 1. 本地配置
localConfig := &config.Config{
    App: appConfig,
}

// 2. 从 Consul 读取配置
consulConfigData := consul.GlobalClient.GetKV("myapp/config")
consulConfig := parseConsulConfig(consulConfigData)

// 3. 合并配置
finalConfig := config.MergeConfig(localConfig, consulConfig)

// 4. 使用合并后的配置
appConfig = &finalConfig.App
```

---

## 设计理念

### 1. 统一入口

所有环境变量通过 `EnvArgs` 统一管理，避免在各个模块中分散读取。

### 2. 配置分层

```
环境变量 (EnvArgs)
    ↓
本地配置 (Config)
    ↓
Consul 配置 (动态)
    ↓
最终配置 (MergeConfig)
```

### 3. 就地合并

`MergeConfig` 直接修改 config1 并返回其引用，避免创建新对象。

---

**最后更新**: 2025-12-17
