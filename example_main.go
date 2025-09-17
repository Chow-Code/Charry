package main

import (
	"context"
	"encoding/json"
	"time"

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
	eventManager := event.NewEventManager(3) // 使用3个worker

	// 启动事件管理器
	if err := eventManager.Start(); err != nil {
		logger.Fatal("启动事件管理器失败", "error", err)
	}
	defer eventManager.Stop()

	// 创建各种事件处理器
	setupEventHandlers(eventManager)

	// 等待一下让系统准备就绪
	time.Sleep(100 * time.Millisecond)

	// 演示各种事件场景
	demonstrateUserRegistration(eventManager)
	demonstrateOrderProcessing(eventManager)
	demonstratePaymentProcessing(eventManager)
	demonstrateCustomEvents(eventManager)

	// 显示统计信息
	showStatistics(eventManager)

	// 等待所有事件处理完成
	time.Sleep(2 * time.Second)

	logger.Info("事件系统示例程序结束")
}

// setupEventHandlers 设置事件处理器
func setupEventHandlers(em *event.EventManager) {
	logger.Info("开始设置事件处理器")

	// 1. 通用日志处理器 - 记录所有事件
	logHandler := event.NewLoggingHandler("info", "[事件日志]")
	em.Subscribe("*", logHandler) // 使用通配符订阅所有事件类型

	// 2. 用户注册事件处理器
	userRegistrationHandlers(em)

	// 3. 订单事件处理器
	orderEventHandlers(em)

	// 4. 支付事件处理器
	paymentEventHandlers(em)

	// 5. 系统事件处理器
	systemEventHandlers(em)

	logger.Info("事件处理器设置完成")
}

// userRegistrationHandlers 用户注册相关事件处理器
func userRegistrationHandlers(em *event.EventManager) {
	// 欢迎邮件处理器
	welcomeEmailHandler := event.NewEmailHandler(
		[]string{"welcome@company.com"},
		"欢迎新用户",
		"user.registered")
	em.Subscribe("user.registered", welcomeEmailHandler)

	// 用户数据保存处理器
	userDbHandler := event.NewDatabaseHandler("users", "user.registered")
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
func orderEventHandlers(em *event.EventManager) {
	// 订单确认邮件
	orderEmailHandler := event.NewEmailHandler(
		[]string{"orders@company.com"},
		"订单确认",
		"order.created", "order.updated", "order.cancelled")
	em.Subscribe("order.created", orderEmailHandler)
	em.Subscribe("order.updated", orderEmailHandler)
	em.Subscribe("order.cancelled", orderEmailHandler)

	// 库存更新处理器 - 使用HTTP调用库存服务
	inventoryHandler := event.NewHTTPHandler(
		"http://inventory-service/api/update",
		"POST",
		"order.created", "order.cancelled")
	em.Subscribe("order.created", inventoryHandler)
	em.Subscribe("order.cancelled", inventoryHandler)

	// 订单数据保存
	orderDbHandler := event.NewDatabaseHandler("orders",
		"order.created", "order.updated", "order.cancelled")
	em.Subscribe("order.created", orderDbHandler)
	em.Subscribe("order.updated", orderDbHandler)
	em.Subscribe("order.cancelled", orderDbHandler)
}

// paymentEventHandlers 支付相关事件处理器
func paymentEventHandlers(em *event.EventManager) {
	// 支付成功后的复合处理 - 使用链式处理器
	paymentSuccessChain := event.NewChainHandler(false, // 不在错误时停止
		event.NewEmailHandler([]string{"payments@company.com"}, "支付成功通知", "payment.completed"),
		event.NewDatabaseHandler("payments", "payment.completed"),
		event.NewHTTPHandler("http://fulfillment-service/api/process", "POST", "payment.completed"),
	)
	em.Subscribe("payment.completed", paymentSuccessChain)

	// 支付失败处理 - 使用异步链式处理器
	paymentFailureChain := event.NewAsyncChainHandler(5*time.Second,
		event.NewEmailHandler([]string{"support@company.com"}, "支付失败警告", "payment.failed"),
		event.NewDatabaseHandler("payment_failures", "payment.failed"),
		event.NewLoggingHandler("error", "[支付失败]"),
	)
	em.Subscribe("payment.failed", paymentFailureChain)
}

// systemEventHandlers 系统事件处理器
func systemEventHandlers(em *event.EventManager) {
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
func demonstrateUserRegistration(em *event.EventManager) {
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
func demonstrateOrderProcessing(em *event.EventManager) {
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
func demonstratePaymentProcessing(em *event.EventManager) {
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
func demonstrateCustomEvents(em *event.EventManager) {
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
func showStatistics(em *event.EventManager) {
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
