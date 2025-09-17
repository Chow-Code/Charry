package event

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"charry/logger"
)

// LoggingHandler 日志处理器 - 将事件记录到日志
type LoggingHandler struct {
	logLevel string
	prefix   string
}

// NewLoggingHandler 创建日志处理器
func NewLoggingHandler(logLevel, prefix string) *LoggingHandler {
	return &LoggingHandler{
		logLevel: logLevel,
		prefix:   prefix,
	}
}

func (h *LoggingHandler) Handle(ctx context.Context, event Event) error {
	eventJson, _ := json.Marshal(event)
	message := fmt.Sprintf("%s 处理事件", h.prefix)

	switch h.logLevel {
	case "debug":
		logger.Debug(message, "event", string(eventJson))
	case "info":
		logger.Info(message, "event", string(eventJson))
	case "warn":
		logger.Warn(message, "event", string(eventJson))
	case "error":
		logger.Error(message, "event", string(eventJson))
	default:
		logger.Info(message, "event", string(eventJson))
	}

	return nil
}

func (h *LoggingHandler) CanHandle(eventType string) bool {
	return true // 日志处理器可以处理所有事件类型
}

// EmailHandler 邮件处理器 - 模拟发送邮件通知
type EmailHandler struct {
	recipients    []string
	subject       string
	eventTypes    []string
	enabledEvents map[string]bool
}

// NewEmailHandler 创建邮件处理器
func NewEmailHandler(recipients []string, subject string, eventTypes ...string) *EmailHandler {
	enabledEvents := make(map[string]bool)
	for _, eventType := range eventTypes {
		enabledEvents[eventType] = true
	}

	return &EmailHandler{
		recipients:    recipients,
		subject:       subject,
		eventTypes:    eventTypes,
		enabledEvents: enabledEvents,
	}
}

func (h *EmailHandler) Handle(ctx context.Context, event Event) error {
	if !h.CanHandle(event.Type) {
		return fmt.Errorf("邮件处理器不支持事件类型: %s", event.Type)
	}

	// 模拟发送邮件
	emailContent := fmt.Sprintf(
		"事件通知\n"+
			"事件ID: %s\n"+
			"事件类型: %s\n"+
			"事件源: %s\n"+
			"事件时间: %s\n"+
			"事件数据: %v\n",
		event.Id, event.Type, event.Source,
		event.Timestamp.Format(time.RFC3339), event.Data)

	logger.Info("模拟发送邮件",
		"recipients", strings.Join(h.recipients, ","),
		"subject", h.subject,
		"eventId", event.Id,
		"eventType", event.Type,
		"content", emailContent)

	// 模拟发送延迟
	time.Sleep(100 * time.Millisecond)

	return nil
}

func (h *EmailHandler) CanHandle(eventType string) bool {
	if len(h.enabledEvents) == 0 {
		return true
	}
	return h.enabledEvents[eventType]
}

// DatabaseHandler 数据库处理器 - 模拟保存事件到数据库
type DatabaseHandler struct {
	tableName     string
	eventTypes    []string
	enabledEvents map[string]bool
}

// NewDatabaseHandler 创建数据库处理器
func NewDatabaseHandler(tableName string, eventTypes ...string) *DatabaseHandler {
	enabledEvents := make(map[string]bool)
	for _, eventType := range eventTypes {
		enabledEvents[eventType] = true
	}

	return &DatabaseHandler{
		tableName:     tableName,
		eventTypes:    eventTypes,
		enabledEvents: enabledEvents,
	}
}

func (h *DatabaseHandler) Handle(ctx context.Context, event Event) error {
	if !h.CanHandle(event.Type) {
		return fmt.Errorf("数据库处理器不支持事件类型: %s", event.Type)
	}

	// 模拟数据库操作
	eventJson, _ := json.Marshal(event)

	logger.Info("模拟保存事件到数据库",
		"table", h.tableName,
		"eventId", event.Id,
		"eventType", event.Type,
		"eventData", string(eventJson))

	// 模拟数据库写入延迟
	time.Sleep(50 * time.Millisecond)

	return nil
}

func (h *DatabaseHandler) CanHandle(eventType string) bool {
	if len(h.enabledEvents) == 0 {
		return true
	}
	return h.enabledEvents[eventType]
}

// HTTPHandler HTTP处理器 - 模拟发送HTTP请求
type HTTPHandler struct {
	url           string
	method        string
	eventTypes    []string
	enabledEvents map[string]bool
}

// NewHTTPHandler 创建HTTP处理器
func NewHTTPHandler(url, method string, eventTypes ...string) *HTTPHandler {
	enabledEvents := make(map[string]bool)
	for _, eventType := range eventTypes {
		enabledEvents[eventType] = true
	}

	return &HTTPHandler{
		url:           url,
		method:        method,
		eventTypes:    eventTypes,
		enabledEvents: enabledEvents,
	}
}

func (h *HTTPHandler) Handle(ctx context.Context, event Event) error {
	if !h.CanHandle(event.Type) {
		return fmt.Errorf("HTTP处理器不支持事件类型: %s", event.Type)
	}

	// 模拟HTTP请求
	eventJson, _ := json.Marshal(event)

	logger.Info("模拟发送HTTP请求",
		"url", h.url,
		"method", h.method,
		"eventId", event.Id,
		"eventType", event.Type,
		"payload", string(eventJson))

	// 模拟网络请求延迟
	time.Sleep(200 * time.Millisecond)

	return nil
}

func (h *HTTPHandler) CanHandle(eventType string) bool {
	if len(h.enabledEvents) == 0 {
		return true
	}
	return h.enabledEvents[eventType]
}

// FunctionHandler 函数处理器 - 使用自定义函数处理事件
type FunctionHandler struct {
	handlerFunc   func(ctx context.Context, event Event) error
	canHandleFunc func(eventType string) bool
	name          string
}

// NewFunctionHandler 创建函数处理器
func NewFunctionHandler(name string, handlerFunc func(ctx context.Context, event Event) error, canHandleFunc func(eventType string) bool) *FunctionHandler {
	return &FunctionHandler{
		handlerFunc:   handlerFunc,
		canHandleFunc: canHandleFunc,
		name:          name,
	}
}

func (h *FunctionHandler) Handle(ctx context.Context, event Event) error {
	if h.handlerFunc == nil {
		return fmt.Errorf("函数处理器 %s 没有设置处理函数", h.name)
	}

	logger.Debug("使用函数处理器处理事件",
		"handlerName", h.name,
		"eventId", event.Id,
		"eventType", event.Type)

	return h.handlerFunc(ctx, event)
}

func (h *FunctionHandler) CanHandle(eventType string) bool {
	if h.canHandleFunc == nil {
		return true
	}
	return h.canHandleFunc(eventType)
}

// ChainHandler 链式处理器 - 按顺序执行多个处理器
type ChainHandler struct {
	handlers    []EventHandler
	stopOnError bool
}

// NewChainHandler 创建链式处理器
func NewChainHandler(stopOnError bool, handlers ...EventHandler) *ChainHandler {
	return &ChainHandler{
		handlers:    handlers,
		stopOnError: stopOnError,
	}
}

func (h *ChainHandler) Handle(ctx context.Context, event Event) error {
	var errors []error

	for i, handler := range h.handlers {
		if !handler.CanHandle(event.Type) {
			continue
		}

		logger.Debug("链式处理器执行子处理器",
			"handlerIndex", i,
			"eventId", event.Id,
			"eventType", event.Type)

		if err := handler.Handle(ctx, event); err != nil {
			errors = append(errors, err)

			logger.Error("链式处理器中的子处理器执行失败",
				"handlerIndex", i,
				"eventId", event.Id,
				"eventType", event.Type,
				"error", err)

			if h.stopOnError {
				return fmt.Errorf("链式处理器在第%d个处理器失败: %v", i, err)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("链式处理器中有%d个处理器失败: %v", len(errors), errors)
	}

	return nil
}

func (h *ChainHandler) CanHandle(eventType string) bool {
	// 只要有一个子处理器能处理，链式处理器就能处理
	for _, handler := range h.handlers {
		if handler.CanHandle(eventType) {
			return true
		}
	}
	return false
}

// AsyncChainHandler 异步链式处理器 - 并发执行多个处理器
type AsyncChainHandler struct {
	handlers []EventHandler
	timeout  time.Duration
}

// NewAsyncChainHandler 创建异步链式处理器
func NewAsyncChainHandler(timeout time.Duration, handlers ...EventHandler) *AsyncChainHandler {
	return &AsyncChainHandler{
		handlers: handlers,
		timeout:  timeout,
	}
}

func (h *AsyncChainHandler) Handle(ctx context.Context, event Event) error {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	errChan := make(chan error, len(h.handlers))
	var handlerCount int

	// 启动所有能处理该事件的处理器
	for i, handler := range h.handlers {
		if !handler.CanHandle(event.Type) {
			continue
		}

		handlerCount++
		go func(index int, h EventHandler) {
			logger.Debug("异步链式处理器执行子处理器",
				"handlerIndex", index,
				"eventId", event.Id,
				"eventType", event.Type)

			err := h.Handle(ctx, event)
			if err != nil {
				logger.Error("异步链式处理器中的子处理器执行失败",
					"handlerIndex", index,
					"eventId", event.Id,
					"eventType", event.Type,
					"error", err)
			}
			errChan <- err
		}(i, handler)
	}

	if handlerCount == 0 {
		return nil
	}

	// 等待所有处理器完成
	var errors []error
	for i := 0; i < handlerCount; i++ {
		select {
		case err := <-errChan:
			if err != nil {
				errors = append(errors, err)
			}
		case <-ctx.Done():
			return fmt.Errorf("异步链式处理器超时: %v", ctx.Err())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("异步链式处理器中有%d个处理器失败: %v", len(errors), errors)
	}

	return nil
}

func (h *AsyncChainHandler) CanHandle(eventType string) bool {
	// 只要有一个子处理器能处理，异步链式处理器就能处理
	for _, handler := range h.handlers {
		if handler.CanHandle(eventType) {
			return true
		}
	}
	return false
}
