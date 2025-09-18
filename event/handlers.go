package event

import (
	"context"
	"fmt"
	"time"

	"charry/logger"
)

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
	handlers    []Handler
	stopOnError bool
}

// NewChainHandler 创建链式处理器
func NewChainHandler(stopOnError bool, handlers ...Handler) *ChainHandler {
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
	handlers []Handler
	timeout  time.Duration
}

// NewAsyncChainHandler 创建异步链式处理器
func NewAsyncChainHandler(timeout time.Duration, handlers ...Handler) *AsyncChainHandler {
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
		go func(index int, h Handler) {
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
