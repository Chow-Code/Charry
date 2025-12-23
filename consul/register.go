package consul

import (
	"encoding/json"
	"fmt"

	"github.com/charry/config"
	consulapi "github.com/hashicorp/consul/api"
)

// RegisterService 将 AppConfig 注册到 Consul
func (c *Client) RegisterService(appConfig *config.AppConfig) error {
	if appConfig == nil {
		return fmt.Errorf("appConfig is nil")
	}

	// 构建服务 ID（唯一标识）
	serviceID := fmt.Sprintf("%s-%s-%d", appConfig.Type, appConfig.Environment, appConfig.Id)

	// 构建服务名称（同类服务共享同一名称）
	serviceName := fmt.Sprintf("%s-%s", appConfig.Type, appConfig.Environment)

	// 构建服务地址
	serviceAddr := appConfig.Addr.Host
	servicePort := appConfig.Addr.Port

	// 构建标签
	tags := []string{
		fmt.Sprintf("id:%d", appConfig.Id),
		fmt.Sprintf("type:%s", appConfig.Type),
		fmt.Sprintf("env:%s", appConfig.Environment),
	}

	// 构建 Metadata（将 AppConfig 展开）
	meta, err := buildMetadata(appConfig)
	if err != nil {
		return fmt.Errorf("构建 Metadata 失败: %w", err)
	}

	// 构建服务注册信息
	registration := &consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Tags:    tags,
		Address: serviceAddr,
		Port:    servicePort,
		Meta:    meta,
		Check:   c.createHealthCheck(serviceAddr, servicePort),
	}

	// 注册服务
	if err := c.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	return nil
}

// DeregisterService 从 Consul 注销服务
func (c *Client) DeregisterService(appConfig *config.AppConfig) error {
	if appConfig == nil {
		return fmt.Errorf("appConfig is nil")
	}

	serviceID := fmt.Sprintf("%s-%s-%d", appConfig.Type, appConfig.Environment, appConfig.Id)

	err := c.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	return nil
}

// GetService 获取服务信息
func (c *Client) GetService(serviceName string) ([]*consulapi.ServiceEntry, error) {
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return services, nil
}

// GetHealthyService 获取健康的服务实例
func (c *Client) GetHealthyService(serviceName string) ([]*consulapi.ServiceEntry, error) {
	// passing=true 表示只返回健康的服务
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy service: %w", err)
	}

	return services, nil
}

// ListServices 列出所有服务
func (c *Client) ListServices() (map[string][]string, error) {
	services, err := c.client.Agent().Services()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	result := make(map[string][]string)
	for _, service := range services {
		result[service.Service] = service.Tags
	}

	return result, nil
}

// buildMetadata 构建 Consul Metadata
// 将 AppConfig 的字段展开到 Meta 中，data 字段转换为 JSON 字符串
func buildMetadata(appConfig *config.AppConfig) (map[string]string, error) {
	meta := make(map[string]string)

	// 1. 先将 AppConfig 序列化为 JSON
	appConfigJSON, err := json.Marshal(appConfig)
	if err != nil {
		return nil, fmt.Errorf("序列化 AppConfig 失败: %w", err)
	}

	// 2. 解析 JSON 到 map
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(appConfigJSON, &jsonMap); err != nil {
		return nil, fmt.Errorf("解析 AppConfig JSON 失败: %w", err)
	}

	// 3. 遍历 JSON 的 key-value
	for key, value := range jsonMap {
		switch key {
		case "data":
			// data 字段特殊处理：转换为 JSON 字符串
			if dataValue, ok := value.(map[string]interface{}); ok && len(dataValue) > 0 {
				dataJSON, err := json.Marshal(dataValue)
				if err == nil {
					meta["data"] = string(dataJSON)
				}
			}
		case "addr":
			// addr 字段特殊处理：展开为 host 和 port
			if addrValue, ok := value.(map[string]interface{}); ok {
				if host, ok := addrValue["host"].(string); ok {
					meta["host"] = host
				}
				if port, ok := addrValue["port"].(float64); ok {
					meta["port"] = fmt.Sprintf("%d", int(port))
				}
			}
		default:
			// 其他字段：直接转换为字符串
			meta[key] = fmt.Sprintf("%v", value)
		}
	}

	return meta, nil
}

// createHealthCheck 根据配置创建健康检查
// 只使用 TCP 端口检查（简单可靠）
func (c *Client) createHealthCheck(addr string, port int) *consulapi.AgentServiceCheck {
	cfg := config.Get()

	return &consulapi.AgentServiceCheck{
		TCP:                            fmt.Sprintf("%s:%d", addr, port),
		Interval:                       cfg.Consul.HealthCheckInterval,
		Timeout:                        cfg.Consul.HealthCheckTimeout,
		DeregisterCriticalServiceAfter: cfg.Consul.DeregisterCriticalServiceAfter,
	}
}

// UpdateHealthCheckTTL 更新 TTL 健康检查状态
// 当使用 TTL 健康检查时，服务需要定期调用此方法报告健康状态
// status 可以是："pass", "warn", "fail"
func (c *Client) UpdateHealthCheckTTL(appConfig *config.AppConfig, status string, output string) error {
	checkID := fmt.Sprintf("service:%s-%s-%d",
		appConfig.Type, appConfig.Environment, appConfig.Id)

	return c.client.Agent().UpdateTTL(checkID, output, status)
}

// PassHealthCheck 标记健康检查为通过（TTL 模式）
func (c *Client) PassHealthCheck(appConfig *config.AppConfig) error {
	return c.UpdateHealthCheckTTL(appConfig, "pass", "Service is healthy")
}

// WarnHealthCheck 标记健康检查为警告（TTL 模式）
func (c *Client) WarnHealthCheck(appConfig *config.AppConfig, reason string) error {
	return c.UpdateHealthCheckTTL(appConfig, "warn", reason)
}

// FailHealthCheck 标记健康检查为失败（TTL 模式）
func (c *Client) FailHealthCheck(appConfig *config.AppConfig, reason string) error {
	return c.UpdateHealthCheckTTL(appConfig, "fail", reason)
}
