package event

import "github.com/charry/logger"

// 示例：用户注册事件消费者
type UserRegisterConsumer struct{}

// CaseEvent 关注的事件
func (c *UserRegisterConsumer) CaseEvent() []string {
	return []string{"user.register", "user.login"}
}

// Triggered 事件触发时执行
func (c *UserRegisterConsumer) Triggered(event *Event) error {
	logger.Infof("用户事件触发: %s", event.Name)

	// 处理事件数据
	if data, ok := event.Data.(map[string]interface{}); ok {
		logger.Infof("事件数据: %+v", data)
	}

	return nil
}

// Async 异步执行
func (c *UserRegisterConsumer) Async() bool {
	return true // 异步执行
}

// 示例：订单支付事件消费者（同步执行）
type OrderPaymentConsumer struct{}

func (c *OrderPaymentConsumer) CaseEvent() []string {
	return []string{"order.payment"}
}

func (c *OrderPaymentConsumer) Triggered(event *Event) error {
	logger.Infof("订单支付事件: %s", event.Name)

	// 同步处理支付逻辑
	// ...

	return nil
}

func (c *OrderPaymentConsumer) Async() bool {
	return false // 同步执行（确保支付顺序）
}

// 使用示例：
//
// func main() {
//     // 1. 初始化事件模块
//     event.Init(10)
//
//     // 2. 注册消费者
//     event.Register(&UserRegisterConsumer{})
//     event.Register(&OrderPaymentConsumer{})
//
//     // 3. 发布事件
//     event.PublishEvent("user.register", map[string]interface{}{
//         "user_id": 123,
//         "username": "john",
//     })
//
//     // 4. 关闭
//     defer event.Close()
// }
