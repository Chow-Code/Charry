package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"charry/cluster"
	"charry/event"
	"charry/logger"
)

// UserRegisteredData 用户注册事件数据
type UserRegisteredData struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// OrderCreatedData 订单创建事件数据
type OrderCreatedData struct {
	OrderId string  `json:"order_id"`
	UserId  string  `json:"user_id"`
	Amount  float64 `json:"amount"`
	Items   []Item  `json:"items"`
}

type Item struct {
	ProductId string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// PaymentProcessedData 支付处理事件数据
type PaymentProcessedData struct {
	OrderId   string  `json:"order_id"`
	UserId    string  `json:"user_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	PaymentId string  `json:"payment_id"`
}

func main() {
	// 初始化日志系统
	logger.Info("启动事件系统示例程序")

	// 创建事件管理器
	eventManager := event.NewManager(3) // 使用3个worker

	// 启动事件管理器
	if err := eventManager.Start(); err != nil {
		logger.Fatal("启动事件管理器失败", "error", err)
	}
	defer eventManager.Stop()

	// 创建各种事件处理器
	setupEventHandlers(eventManager)

	// 等待一下让系统准备就绪
	time.Sleep(100 * time.Millisecond)

	// 创建集群管理器（可选，用于演示集群功能）
	clusterManager := setupClusterManager(eventManager)
	defer func() {
		if clusterManager != nil {
			clusterManager.Stop()
		}
	}()

	// 演示各种事件场景
	demonstrateUserRegistration(eventManager)
	demonstrateOrderProcessing(eventManager)
	demonstratePaymentProcessing(eventManager)
	demonstrateCustomEvents(eventManager)

	// 演示集群管理功能（如果集群管理器可用）
	if clusterManager != nil {
		demonstrateClusterManagement(clusterManager)
	}

	// 显示统计信息
	showStatistics(eventManager)

	// 等待所有事件处理完成
	time.Sleep(2 * time.Second)

	logger.Info("事件系统示例程序结束")
}

// setupEventHandlers 设置事件处理器
func setupEventHandlers(em *event.Manager) {
	logger.Info("开始设置事件处理器")

	// 1. 通用日志处理器 - 使用函数处理器记录所有事件
	logHandler := event.NewFunctionHandler(
		"通用日志处理器",
		func(ctx context.Context, event event.Event) error {
			eventJson, _ := json.Marshal(event)
			logger.Info("[事件日志] 处理事件", "event", string(eventJson))
			return nil
		},
		func(eventType string) bool {
			return true // 处理所有事件类型
		},
	)
	em.Subscribe("*", logHandler) // 使用通配符订阅所有事件类型

	// 2. 用户注册事件处理器
	userRegistrationHandlers(em)

	// 3. 订单事件处理器
	orderEventHandlers(em)

	// 4. 支付事件处理器
	paymentEventHandlers(em)

	// 5. 系统事件处理器
	systemEventHandlers(em)

	// 6. 集群事件处理器
	clusterEventHandlers(em)

	logger.Info("事件处理器设置完成")
}

// userRegistrationHandlers 用户注册相关事件处理器
func userRegistrationHandlers(em *event.Manager) {
	// 欢迎邮件处理器 - 使用函数处理器模拟发送邮件
	welcomeEmailHandler := event.NewFunctionHandler(
		"欢迎邮件处理器",
		func(ctx context.Context, event event.Event) error {
			logger.Info("模拟发送欢迎邮件",
				"eventId", event.Id,
				"eventType", event.Type,
				"recipients", "welcome@company.com")

			// 模拟邮件发送延迟
			time.Sleep(100 * time.Millisecond)
			return nil
		},
		func(eventType string) bool {
			return eventType == "user.registered"
		},
	)
	em.Subscribe("user.registered", welcomeEmailHandler)

	// 用户数据保存处理器 - 使用函数处理器模拟数据库操作
	userDbHandler := event.NewFunctionHandler(
		"用户数据保存处理器",
		func(ctx context.Context, event event.Event) error {
			eventJson, _ := json.Marshal(event)
			logger.Info("模拟保存用户数据到数据库",
				"eventId", event.Id,
				"eventType", event.Type,
				"table", "users",
				"eventData", string(eventJson))

			// 模拟数据库写入延迟
			time.Sleep(50 * time.Millisecond)
			return nil
		},
		func(eventType string) bool {
			return eventType == "user.registered"
		},
	)
	em.Subscribe("user.registered", userDbHandler)

	// 用户统计处理器 - 使用自定义函数
	userStatsHandler := event.NewFunctionHandler(
		"用户统计处理器",
		func(ctx context.Context, event event.Event) error {
			logger.Info("更新用户统计数据",
				"eventId", event.Id,
				"eventType", event.Type)
			// 这里可以实现实际的统计逻辑
			return nil
		},
		func(eventType string) bool {
			return eventType == "user.registered" || eventType == "user.deleted"
		},
	)
	em.Subscribe("user.registered", userStatsHandler)
	em.Subscribe("user.deleted", userStatsHandler)
}

// orderEventHandlers 订单相关事件处理器
func orderEventHandlers(em *event.Manager) {
	// 订单确认邮件处理器
	orderEmailHandler := event.NewFunctionHandler(
		"订单确认邮件处理器",
		func(ctx context.Context, event event.Event) error {
			logger.Info("模拟发送订单确认邮件",
				"eventId", event.Id,
				"eventType", event.Type,
				"recipients", "orders@company.com")

			time.Sleep(100 * time.Millisecond)
			return nil
		},
		func(eventType string) bool {
			return eventType == "order.created" || eventType == "order.updated" || eventType == "order.cancelled"
		},
	)
	em.Subscribe("order.created", orderEmailHandler)
	em.Subscribe("order.updated", orderEmailHandler)
	em.Subscribe("order.cancelled", orderEmailHandler)

	// 库存更新处理器 - 使用函数处理器模拟HTTP调用
	inventoryHandler := event.NewFunctionHandler(
		"库存更新处理器",
		func(ctx context.Context, event event.Event) error {
			eventJson, _ := json.Marshal(event)
			logger.Info("模拟发送HTTP请求到库存服务",
				"eventId", event.Id,
				"eventType", event.Type,
				"url", "http://inventory-service/api/update",
				"method", "POST",
				"payload", string(eventJson))

			// 模拟网络请求延迟
			time.Sleep(200 * time.Millisecond)
			return nil
		},
		func(eventType string) bool {
			return eventType == "order.created" || eventType == "order.cancelled"
		},
	)
	em.Subscribe("order.created", inventoryHandler)
	em.Subscribe("order.cancelled", inventoryHandler)

	// 订单数据保存处理器
	orderDbHandler := event.NewFunctionHandler(
		"订单数据保存处理器",
		func(ctx context.Context, event event.Event) error {
			eventJson, _ := json.Marshal(event)
			logger.Info("模拟保存订单数据到数据库",
				"eventId", event.Id,
				"eventType", event.Type,
				"table", "orders",
				"eventData", string(eventJson))

			time.Sleep(50 * time.Millisecond)
			return nil
		},
		func(eventType string) bool {
			return eventType == "order.created" || eventType == "order.updated" || eventType == "order.cancelled"
		},
	)
	em.Subscribe("order.created", orderDbHandler)
	em.Subscribe("order.updated", orderDbHandler)
	em.Subscribe("order.cancelled", orderDbHandler)
}

// paymentEventHandlers 支付相关事件处理器
func paymentEventHandlers(em *event.Manager) {
	// 支付成功后的复合处理 - 使用链式处理器
	paymentEmailHandler := event.NewFunctionHandler(
		"支付成功邮件处理器",
		func(ctx context.Context, event event.Event) error {
			logger.Info("模拟发送支付成功邮件",
				"eventId", event.Id,
				"eventType", event.Type,
				"recipients", "payments@company.com")

			time.Sleep(100 * time.Millisecond)
			return nil
		},
		func(eventType string) bool {
			return eventType == "payment.completed"
		},
	)

	paymentDbHandler := event.NewFunctionHandler(
		"支付数据保存处理器",
		func(ctx context.Context, event event.Event) error {
			eventJson, _ := json.Marshal(event)
			logger.Info("模拟保存支付数据到数据库",
				"eventId", event.Id,
				"eventType", event.Type,
				"table", "payments",
				"eventData", string(eventJson))

			time.Sleep(50 * time.Millisecond)
			return nil
		},
		func(eventType string) bool {
			return eventType == "payment.completed"
		},
	)

	fulfillmentHandler := event.NewFunctionHandler(
		"订单履约处理器",
		func(ctx context.Context, event event.Event) error {
			eventJson, _ := json.Marshal(event)
			logger.Info("模拟发送HTTP请求到履约服务",
				"eventId", event.Id,
				"eventType", event.Type,
				"url", "http://fulfillment-service/api/process",
				"method", "POST",
				"payload", string(eventJson))

			time.Sleep(200 * time.Millisecond)
			return nil
		},
		func(eventType string) bool {
			return eventType == "payment.completed"
		},
	)

	paymentSuccessChain := event.NewChainHandler(false, // 不在错误时停止
		paymentEmailHandler,
		paymentDbHandler,
		fulfillmentHandler,
	)
	em.Subscribe("payment.completed", paymentSuccessChain)

	// 支付失败处理 - 使用异步链式处理器
	failureEmailHandler := event.NewFunctionHandler(
		"支付失败邮件处理器",
		func(ctx context.Context, event event.Event) error {
			logger.Info("模拟发送支付失败告警邮件",
				"eventId", event.Id,
				"eventType", event.Type,
				"recipients", "support@company.com")

			time.Sleep(100 * time.Millisecond)
			return nil
		},
		func(eventType string) bool {
			return eventType == "payment.failed"
		},
	)

	failureDbHandler := event.NewFunctionHandler(
		"支付失败数据保存处理器",
		func(ctx context.Context, event event.Event) error {
			eventJson, _ := json.Marshal(event)
			logger.Info("模拟保存支付失败数据到数据库",
				"eventId", event.Id,
				"eventType", event.Type,
				"table", "payment_failures",
				"eventData", string(eventJson))

			time.Sleep(50 * time.Millisecond)
			return nil
		},
		func(eventType string) bool {
			return eventType == "payment.failed"
		},
	)

	failureLogHandler := event.NewFunctionHandler(
		"支付失败日志处理器",
		func(ctx context.Context, event event.Event) error {
			eventJson, _ := json.Marshal(event)
			logger.Error("[支付失败] 处理事件", "event", string(eventJson))
			return nil
		},
		func(eventType string) bool {
			return eventType == "payment.failed"
		},
	)

	paymentFailureChain := event.NewAsyncChainHandler(5*time.Second,
		failureEmailHandler,
		failureDbHandler,
		failureLogHandler,
	)
	em.Subscribe("payment.failed", paymentFailureChain)
}

// systemEventHandlers 系统事件处理器
func systemEventHandlers(em *event.Manager) {
	// 系统错误处理器 - 只处理严重错误
	errorFilter := func(e event.Event) bool {
		if metadata, ok := e.Metadata["severity"]; ok {
			return metadata == "critical" || metadata == "high"
		}
		return true
	}

	systemErrorHandler := event.NewFunctionHandler(
		"系统错误处理器",
		func(ctx context.Context, event event.Event) error {
			logger.Error("系统发生严重错误",
				"eventId", event.Id,
				"eventType", event.Type,
				"eventData", event.Data)

			// 这里可以实现告警逻辑，比如发送钉钉消息、短信等
			return nil
		},
		func(eventType string) bool {
			return eventType == "system.error"
		},
	)

	// 使用过滤器订阅
	subId, _ := em.Subscribe("system.error", systemErrorHandler, errorFilter)
	logger.Info("订阅系统错误事件", "subscriptionId", subId)
}

// demonstrateUserRegistration 演示用户注册事件
func demonstrateUserRegistration(em *event.Manager) {
	logger.Info("=== 演示用户注册事件 ===")

	userData := UserRegisteredData{
		UserId:   "user_001",
		Username: "张三",
		Email:    "zhangsan@example.com",
	}

	userEvent := event.NewEvent("user.registered", "user-service", userData).
		WithMetadata("ip", "192.168.1.100").
		WithMetadata("user_agent", "Chrome/91.0")

	if err := em.Publish(userEvent); err != nil {
		logger.Error("发布用户注册事件失败", "error", err)
	} else {
		logger.Info("用户注册事件已发布", "eventId", userEvent.Id)
	}
}

// demonstrateOrderProcessing 演示订单处理事件
func demonstrateOrderProcessing(em *event.Manager) {
	logger.Info("=== 演示订单处理事件 ===")

	// 创建订单事件
	orderData := OrderCreatedData{
		OrderId: "order_001",
		UserId:  "user_001",
		Amount:  199.99,
		Items: []Item{
			{ProductId: "prod_001", Quantity: 2, Price: 99.99},
		},
	}

	orderEvent := event.NewEvent("order.created", "order-service", orderData).
		WithMetadata("channel", "web").
		WithMetadata("promotion_code", "WELCOME10")

	if err := em.Publish(orderEvent); err != nil {
		logger.Error("发布订单创建事件失败", "error", err)
	}

	// 等待一段时间后更新订单
	time.Sleep(500 * time.Millisecond)

	orderData.Amount = 179.99 // 应用折扣
	updateEvent := event.NewEvent("order.updated", "order-service", orderData).
		WithMetadata("update_reason", "discount_applied")

	if err := em.Publish(updateEvent); err != nil {
		logger.Error("发布订单更新事件失败", "error", err)
	}
}

// demonstratePaymentProcessing 演示支付处理事件
func demonstratePaymentProcessing(em *event.Manager) {
	logger.Info("=== 演示支付处理事件 ===")

	// 支付成功事件
	paymentData := PaymentProcessedData{
		OrderId:   "order_001",
		UserId:    "user_001",
		Amount:    179.99,
		Status:    "completed",
		PaymentId: "pay_001",
	}

	paymentEvent := event.NewEvent("payment.completed", "payment-service", paymentData).
		WithMetadata("payment_method", "credit_card").
		WithMetadata("gateway", "stripe")

	if err := em.Publish(paymentEvent); err != nil {
		logger.Error("发布支付完成事件失败", "error", err)
	}

	// 等待一段时间后模拟一个支付失败事件
	time.Sleep(300 * time.Millisecond)

	failedPaymentData := PaymentProcessedData{
		OrderId:   "order_002",
		UserId:    "user_002",
		Amount:    299.99,
		Status:    "failed",
		PaymentId: "pay_002",
	}

	failedPaymentEvent := event.NewEvent("payment.failed", "payment-service", failedPaymentData).
		WithMetadata("payment_method", "credit_card").
		WithMetadata("error_code", "insufficient_funds")

	if err := em.Publish(failedPaymentEvent); err != nil {
		logger.Error("发布支付失败事件失败", "error", err)
	}
}

// demonstrateCustomEvents 演示自定义事件
func demonstrateCustomEvents(em *event.Manager) {
	logger.Info("=== 演示自定义事件 ===")

	// 系统错误事件（高优先级）
	systemErrorEvent := event.NewEvent("system.error", "api-gateway", map[string]interface{}{
		"error":   "Database connection timeout",
		"service": "order-service",
		"impact":  "Service temporarily unavailable",
	}).WithMetadata("severity", "critical")

	if err := em.Publish(systemErrorEvent); err != nil {
		logger.Error("发布系统错误事件失败", "error", err)
	}

	// 系统错误事件（低优先级，会被过滤器过滤掉）
	lowPriorityErrorEvent := event.NewEvent("system.error", "monitoring", map[string]interface{}{
		"error":   "Minor cache miss",
		"service": "cache-service",
	}).WithMetadata("severity", "low")

	if err := em.Publish(lowPriorityErrorEvent); err != nil {
		logger.Error("发布低优先级系统错误事件失败", "error", err)
	}

	// 自定义业务事件
	customEvent := event.NewEvent("business.milestone", "analytics-service", map[string]interface{}{
		"milestone": "1000_orders",
		"value":     1000,
		"message":   "系统已处理1000个订单!",
	})

	if err := em.Publish(customEvent); err != nil {
		logger.Error("发布自定义业务事件失败", "error", err)
	}
}

// showStatistics 显示统计信息
func showStatistics(em *event.Manager) {
	logger.Info("=== 事件管理器统计信息 ===")

	stats := em.GetStats()
	statsJson, _ := json.MarshalIndent(stats, "", "  ")
	logger.Info("统计信息", "stats", string(statsJson))

	subscriptions := em.GetSubscriptions()
	logger.Info("订阅信息", "总订阅数", func() int {
		count := 0
		for _, subs := range subscriptions {
			count += len(subs)
		}
		return count
	}())

	for eventType, subs := range subscriptions {
		logger.Info("事件类型订阅详情",
			"eventType", eventType,
			"subscriberCount", len(subs))
	}
}

// clusterEventHandlers 集群相关事件处理器
func clusterEventHandlers(em *event.Manager) {
	logger.Info("设置集群事件处理器")

	// 节点变化通知处理器
	nodeChangeHandler := event.NewFunctionHandler(
		"节点变化通知处理器",
		func(ctx context.Context, event event.Event) error {
			logger.Info("检测到节点变化",
				"eventId", event.Id,
				"eventType", event.Type,
				"source", event.Source)

			// 根据事件类型处理
			switch event.Type {
			case cluster.EventNodeAdded:
				if nodeData, ok := event.Data.(*cluster.NodeEventData); ok {
					logger.Info("节点已添加",
						"nodeId", nodeData.Node.Id,
						"nodeName", nodeData.Node.Name,
						"nodeType", nodeData.Node.Type,
						"lanAddr", nodeData.Node.LanAddr,
						"reason", nodeData.Reason)
				}
			case cluster.EventNodeUpdated:
				if nodeData, ok := event.Data.(*cluster.NodeEventData); ok {
					logger.Info("节点已更新",
						"nodeId", nodeData.Node.Id,
						"nodeName", nodeData.Node.Name,
						"reason", nodeData.Reason)
				}
			case cluster.EventNodeRemoved:
				if nodeData, ok := event.Data.(*cluster.NodeEventData); ok {
					logger.Info("节点已删除",
						"nodeId", nodeData.Node.Id,
						"nodeName", nodeData.Node.Name,
						"reason", nodeData.Reason)
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
	em.Subscribe(cluster.EventNodeAdded, nodeChangeHandler)
	em.Subscribe(cluster.EventNodeUpdated, nodeChangeHandler)
	em.Subscribe(cluster.EventNodeRemoved, nodeChangeHandler)

	// 集群整体变化处理器
	clusterChangeHandler := event.NewFunctionHandler(
		"集群变化处理器",
		func(ctx context.Context, event event.Event) error {
			logger.Info("检测到集群变化",
				"eventId", event.Id,
				"eventType", event.Type)

			if clusterData, ok := event.Data.(*cluster.ClusterEventData); ok {
				logger.Info("集群状态统计",
					"totalNodes", clusterData.TotalCount,
					"addedCount", len(clusterData.AddedNodes),
					"updatedCount", len(clusterData.UpdatedNodes),
					"removedCount", len(clusterData.RemovedNodes))
			}

			return nil
		},
		func(eventType string) bool {
			return eventType == cluster.EventClusterChanged
		},
	)

	em.Subscribe(cluster.EventClusterChanged, clusterChangeHandler)

	// 集群连接状态处理器
	connectionHandler := event.NewFunctionHandler(
		"集群连接状态处理器",
		func(ctx context.Context, event event.Event) error {
			switch event.Type {
			case cluster.EventClusterConnected:
				logger.Info("集群已连接", "eventId", event.Id)
				if clusterData, ok := event.Data.(*cluster.ClusterEventData); ok {
					logger.Info("集群连接信息", "nodeCount", clusterData.TotalCount)
				}
			case cluster.EventClusterDisconnected:
				logger.Warn("集群连接断开", "eventId", event.Id, "data", event.Data)
			}
			return nil
		},
		func(eventType string) bool {
			return eventType == cluster.EventClusterConnected ||
				eventType == cluster.EventClusterDisconnected
		},
	)

	em.Subscribe(cluster.EventClusterConnected, connectionHandler)
	em.Subscribe(cluster.EventClusterDisconnected, connectionHandler)

	logger.Info("集群事件处理器设置完成")
}

// setupClusterManager 设置集群管理器（注意：需要 Nacos 服务运行）
func setupClusterManager(eventManager *event.Manager) *cluster.Manager {
	logger.Info("=== 初始化集群管理器 ===")

	// 使用默认配置（可以根据实际环境修改）
	config := cluster.DefaultNacosConfig()

	// 可以根据环境变量或配置文件修改配置
	// 这里使用默认的本地 Nacos 配置
	logger.Info("Nacos 配置信息",
		"server", fmt.Sprintf("%s:%d", config.ServerConfigs[0].IpAddr, config.ServerConfigs[0].Port),
		"dataId", config.ClusterConfig.DataId,
		"group", config.ClusterConfig.Group)

	// 创建集群管理器
	clusterManager := cluster.NewManager(config, eventManager)

	// 尝试启动集群管理器
	if err := clusterManager.Start(); err != nil {
		logger.Error("启动集群管理器失败（可能 Nacos 服务未运行）", "error", err)
		logger.Info("跳过集群管理功能演示，继续其他功能")
		return nil
	}

	logger.Info("集群管理器启动成功")
	return clusterManager
}

// demonstrateClusterManagement 演示集群管理功能
func demonstrateClusterManagement(clusterManager *cluster.Manager) {
	logger.Info("=== 演示集群管理功能 ===")

	// 等待一下让系统稳定
	time.Sleep(1 * time.Second)

	// 获取当前所有节点
	currentNodes := clusterManager.GetAllNodes()
	logger.Info("当前集群节点数量", "count", len(currentNodes))

	// 添加测试节点
	testNode1 := &cluster.Node{
		Id:      "node-web-01",
		Name:    "Web服务器-01",
		Type:    1, // Web 服务器
		LanAddr: "192.168.1.10:8080",
		WanAddr: "203.0.113.10:8080",
		Weights: 100,
		Data: map[string]interface{}{
			"cpu_cores": 4,
			"memory_gb": 8,
			"region":    "beijing",
		},
	}

	testNode2 := &cluster.Node{
		Id:      "node-db-01",
		Name:    "数据库服务器-01",
		Type:    2, // 数据库服务器
		LanAddr: "192.168.1.20:3306",
		WanAddr: "203.0.113.20:3306",
		Weights: 150,
		Data: map[string]interface{}{
			"cpu_cores":  8,
			"memory_gb":  32,
			"storage_gb": 500,
			"region":     "beijing",
		},
	}

	// 演示添加节点
	logger.Info("添加测试节点 1")
	if err := clusterManager.AddNode(testNode1); err != nil {
		logger.Error("添加节点失败", "error", err)
	} else {
		logger.Info("节点添加请求已发送，等待 Nacos 同步...")
		time.Sleep(2 * time.Second) // 等待 Nacos 同步和事件处理
	}

	logger.Info("添加测试节点 2")
	if err := clusterManager.AddNode(testNode2); err != nil {
		logger.Error("添加节点失败", "error", err)
	} else {
		logger.Info("节点添加请求已发送，等待 Nacos 同步...")
		time.Sleep(2 * time.Second)
	}

	// 查看更新后的节点列表
	updatedNodes := clusterManager.GetAllNodes()
	logger.Info("更新后的集群节点数量", "count", len(updatedNodes))

	// 演示按类型查询节点
	webNodes := clusterManager.GetNodesByType(1)
	dbNodes := clusterManager.GetNodesByType(2)
	logger.Info("节点类型统计", "webNodes", len(webNodes), "dbNodes", len(dbNodes))

	// 演示更新节点
	if node, exists := clusterManager.GetNode("node-web-01"); exists {
		logger.Info("更新测试节点权重")
		node.Weights = 200
		if err := clusterManager.UpdateNode(node); err != nil {
			logger.Error("更新节点失败", "error", err)
		} else {
			logger.Info("节点更新请求已发送，等待 Nacos 同步...")
			time.Sleep(2 * time.Second)
		}
	}

	// 演示删除节点
	logger.Info("删除测试节点")
	if err := clusterManager.RemoveNode("node-web-01"); err != nil {
		logger.Error("删除节点失败", "error", err)
	} else {
		logger.Info("节点删除请求已发送，等待 Nacos 同步...")
		time.Sleep(2 * time.Second)
	}

	// 最终状态
	finalNodes := clusterManager.GetAllNodes()
	logger.Info("最终集群节点数量", "count", len(finalNodes))

	// 显示集群统计信息
	stats := clusterManager.GetStats()
	statsJson, _ := json.MarshalIndent(stats, "", "  ")
	logger.Info("集群管理器统计信息", "stats", string(statsJson))

	logger.Info("集群管理功能演示完成")
}
