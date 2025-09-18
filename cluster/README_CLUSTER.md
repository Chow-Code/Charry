# 集群管理功能说明

本文档介绍如何使用基于 Nacos 的集群管理功能。

## 功能概述

集群管理功能通过 Nacos 配置中心实现节点的统一管理，支持：

- 📡 **配置拉取**：从 Nacos 拉取节点配置信息
- 🔄 **动态更新**：实时监听配置变化，自动同步节点状态
- 📢 **事件通知**：当节点增加、修改、删除时自动发布事件
- 🏷️ **节点分类**：支持按节点类型查询和管理
- 📊 **状态监控**：提供集群状态统计和监控功能

## 架构组件

### 核心组件

1. **Node**：节点数据结构
2. **NacosConfig**：Nacos 连接配置
3. **Manager**：集群管理器
4. **Events**：集群相关事件定义

### 事件类型

- `cluster.node.added`：节点添加事件
- `cluster.node.updated`：节点更新事件  
- `cluster.node.removed`：节点删除事件
- `cluster.changed`：集群整体变化事件
- `cluster.connected`：集群连接事件
- `cluster.disconnected`：集群断开事件

## 快速开始

### 1. 启动 Nacos 服务

```bash
# 下载并启动 Nacos (单机模式)
wget https://github.com/alibaba/nacos/releases/download/2.2.3/nacos-server-2.2.3.tar.gz
tar -xvf nacos-server-2.2.3.tar.gz
cd nacos/bin
./startup.sh -m standalone

# 访问控制台: http://localhost:8848/nacos
# 默认用户名/密码: nacos/nacos
```

### 2. 基本使用示例

```go
package main

import (
    "charry/cluster"
    "charry/event"
    "charry/logger"
)

func main() {
    // 创建事件管理器
    eventManager := event.NewEventManager(3)
    eventManager.Start()
    defer eventManager.Stop()

    // 创建集群配置
    config := cluster.DefaultNacosConfig()
    
    // 可以自定义配置
    config.ServerConfigs[0].IpAddr = "127.0.0.1"
    config.ServerConfigs[0].Port = 8848
    config.ClusterConfig.DataId = "my-cluster-nodes"
    config.ClusterConfig.Group = "DEFAULT_GROUP"

    // 创建并启动集群管理器
    clusterManager := cluster.NewManager(config, eventManager)
    if err := clusterManager.Start(); err != nil {
        logger.Error("启动集群管理器失败", "error", err)
        return
    }
    defer clusterManager.Stop()

    // 添加节点
    node := &cluster.Node{
        Id:      "web-server-01",
        Name:    "Web服务器01",
        Type:    1,
        LanAddr: "192.168.1.10:8080",
        WanAddr: "203.0.113.10:8080",
        Weights: 100,
        Data: map[string]interface{}{
            "region": "beijing",
            "cpu_cores": 4,
            "memory_gb": 8,
        },
    }

    if err := clusterManager.AddNode(node); err != nil {
        logger.Error("添加节点失败", "error", err)
    }

    // 查询节点
    if node, exists := clusterManager.GetNode("web-server-01"); exists {
        logger.Info("找到节点", "node", node)
    }

    // 按类型查询
    webNodes := clusterManager.GetNodesByType(1)
    logger.Info("Web服务器节点", "count", len(webNodes))
}
```

### 3. 事件订阅示例

```go
// 订阅节点变化事件
nodeChangeHandler := event.NewFunctionHandler(
    "节点变化处理器",
    func(ctx context.Context, event event.Event) error {
        switch event.Type {
        case cluster.EventNodeAdded:
            if nodeData, ok := event.Data.(*cluster.NodeEventData); ok {
                logger.Info("节点已添加", 
                    "nodeId", nodeData.Node.Id,
                    "nodeName", nodeData.Node.Name)
            }
        case cluster.EventNodeUpdated:
            if nodeData, ok := event.Data.(*cluster.NodeEventData); ok {
                logger.Info("节点已更新", 
                    "nodeId", nodeData.Node.Id)
            }
        case cluster.EventNodeRemoved:
            if nodeData, ok := event.Data.(*cluster.NodeEventData); ok {
                logger.Info("节点已删除", 
                    "nodeId", nodeData.Node.Id)
            }
        }
        return nil
    },
    func(eventType string) bool {
        return eventType == cluster.EventNodeAdded ||
               eventType == cluster.EventNodeUpdated ||
               eventType == cluster.EventNodeRemoved
    },
)

// 订阅所有节点事件
eventManager.Subscribe(cluster.EventNodeAdded, nodeChangeHandler)
eventManager.Subscribe(cluster.EventNodeUpdated, nodeChangeHandler)
eventManager.Subscribe(cluster.EventNodeRemoved, nodeChangeHandler)
```

## API 参考

### Node 结构

```go
type Node struct {
    Id      string      // 节点唯一标识
    Name    string      // 节点名称
    Type    int         // 节点类型
    LanAddr string      // 内网地址
    WanAddr string      // 公网地址
    Weights int         // 权重
    Data    interface{} // 节点附加数据
}
```

### Manager 主要方法

#### 节点管理
- `GetAllNodes() []*Node`：获取所有节点
- `GetNode(nodeId string) (*Node, bool)`：根据ID获取节点
- `GetNodesByType(nodeType int) []*Node`：根据类型获取节点
- `AddNode(node *Node) error`：添加节点
- `UpdateNode(node *Node) error`：更新节点
- `RemoveNode(nodeId string) error`：删除节点

#### 配置管理
- `PublishNodeConfig(nodes []*Node) error`：发布节点配置到 Nacos

#### 状态监控
- `GetStats() map[string]interface{}`：获取集群统计信息

### 配置选项

#### NacosConfig 主要配置

```go
type NacosConfig struct {
    ServerConfigs []ServerConfig // Nacos 服务器配置
    ClientConfig  ClientConfig   // 客户端配置
    ClusterConfig ClusterConfig  // 集群配置
}
```

#### 默认配置说明

```go
config := cluster.DefaultNacosConfig()
// 默认配置：
// - Nacos服务器: 127.0.0.1:8848
// - DataId: "cluster-nodes"
// - Group: "DEFAULT_GROUP"
// - 监控间隔: 5秒
```

## 高级功能

### 1. 自定义配置

```go
config := &cluster.NacosConfig{
    ServerConfigs: []cluster.ServerConfig{
        {
            IpAddr: "nacos.example.com",
            Port:   8848,
            ContextPath: "/nacos",
            Scheme: "http",
        },
    },
    ClientConfig: cluster.ClientConfig{
        NamespaceId: "prod",
        Username:    "admin",
        Password:    "password123",
        TimeoutMs:   10000,
    },
    ClusterConfig: cluster.ClusterConfig{
        DataId: "production-cluster",
        Group:  "PROD_GROUP",
        WatchInterval: time.Second * 10,
    },
}
```

### 2. 节点类型定义

建议为不同类型的节点定义常量：

```go
const (
    NodeTypeWeb      = 1  // Web服务器
    NodeTypeDatabase = 2  // 数据库服务器
    NodeTypeCache    = 3  // 缓存服务器
    NodeTypeMQ       = 4  // 消息队列
    NodeTypeAPI      = 5  // API网关
)
```

### 3. 节点健康检查

可以扩展 Node 结构包含健康状态：

```go
type HealthStatus struct {
    Status    string    `json:"status"`     // healthy, unhealthy, unknown
    LastCheck time.Time `json:"last_check"`
    Message   string    `json:"message"`
}

// 在 Node.Data 中存储健康状态
node.Data = map[string]interface{}{
    "health": HealthStatus{
        Status:    "healthy",
        LastCheck: time.Now(),
        Message:   "All services running",
    },
}
```

## 运行示例

运行完整的示例程序：

```bash
# 确保 Nacos 服务已启动
# 然后运行示例
cd /path/to/charry
go run example_main.go
```

示例程序将演示：
- 事件系统基本功能
- 集群管理功能（如果 Nacos 可用）
- 节点的添加、更新、删除
- 事件的发布和处理

## 故障排除

### 1. 连接失败

```
启动集群管理器失败: 获取 Nacos 配置失败: read config from both server and cache fail
```

**解决方案：**
- 确保 Nacos 服务已启动
- 检查网络连接
- 验证 Nacos 地址和端口配置

### 2. 认证失败

**解决方案：**
- 检查用户名密码配置
- 确保 Nacos 启用了认证
- 验证命名空间ID

### 3. 配置不同步

**解决方案：**
- 检查 DataId 和 Group 配置
- 确保配置格式为有效 JSON
- 查看 Nacos 控制台中的配置内容

## 最佳实践

1. **节点命名规范**：使用有意义的节点ID，如 `web-01`、`db-master` 等
2. **类型分类**：合理定义节点类型，便于管理和查询
3. **权重设置**：根据节点性能设置合适的权重值
4. **数据结构**：在 Node.Data 中存储必要的元数据
5. **事件处理**：及时处理节点变化事件，更新本地状态
6. **错误处理**：妥善处理网络异常和配置错误
7. **监控告警**：监控集群连接状态，设置告警机制

## 注意事项

- Nacos 配置变更有一定延迟，通常在几秒内同步
- 大量节点变更时建议批量操作而非频繁单次操作
- 生产环境建议使用 Nacos 集群部署保证高可用
- 定期备份 Nacos 配置数据
