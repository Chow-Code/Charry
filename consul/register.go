package consul

import (
	"fmt"
	"strings"

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

	// 从 Metadata 中添加额外标签
	for key, value := range appConfig.Metadata {
		tags = append(tags, fmt.Sprintf("%s:%v", key, value))
	}

	// 构建服务注册信息
	registration := &consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Tags:    tags,
		Address: serviceAddr,
		Port:    servicePort,
		Meta:    convertMetadata(appConfig.Metadata),
		Check:   c.createHealthCheck(serviceAddr, servicePort),
	}

	// 注册服务
	err := c.client.Agent().ServiceRegister(registration)
	if err != nil {
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

// convertMetadata 转换 metadata 为 map[string]string
func convertMetadata(metadata map[string]any) map[string]string {
	result := make(map[string]string)
	for key, value := range metadata {
		result[key] = fmt.Sprintf("%v", value)
	}
	return result
}

// createHealthCheck 根据配置创建健康检查
func (c *Client) createHealthCheck(addr string, port int) *consulapi.AgentServiceCheck {
	switch c.config.HealthCheckType {
	case HealthCheckTypeHTTP:
		// HTTP 健康检查
		path := c.config.HealthCheckPath
		if path == "" {
			path = "/health"
		}
		url := fmt.Sprintf("http://%s:%d%s", addr, port, path)
		return &consulapi.AgentServiceCheck{
			HTTP:                           url,
			Interval:                       c.config.HealthCheckInterval,
			Timeout:                        c.config.HealthCheckTimeout,
			DeregisterCriticalServiceAfter: c.config.DeregisterCriticalServiceAfter,
		}

	case HealthCheckTypeGRPC:
		// gRPC 健康检查（使用 gRPC 健康检查协议）
		// 格式：host:port[/service]
		grpcAddr := fmt.Sprintf("%s:%d", addr, port)
		if c.config.HealthCheckPath != "" {
			// 移除前导斜杠
			service := strings.TrimPrefix(c.config.HealthCheckPath, "/")
			grpcAddr = fmt.Sprintf("%s/%s", grpcAddr, service)
		}
		return &consulapi.AgentServiceCheck{
			GRPC:                           grpcAddr,
			GRPCUseTLS:                     c.config.GRPCUseTLS,
			Interval:                       c.config.HealthCheckInterval,
			Timeout:                        c.config.HealthCheckTimeout,
			DeregisterCriticalServiceAfter: c.config.DeregisterCriticalServiceAfter,
		}

	case HealthCheckTypeTCP:
		// TCP 健康检查（只检查端口是否可达）
		tcpAddr := fmt.Sprintf("%s:%d", addr, port)
		return &consulapi.AgentServiceCheck{
			TCP:                            tcpAddr,
			Interval:                       c.config.HealthCheckInterval,
			Timeout:                        c.config.HealthCheckTimeout,
			DeregisterCriticalServiceAfter: c.config.DeregisterCriticalServiceAfter,
		}

	case HealthCheckTypeTTL:
		// TTL 健康检查（服务自己定期报告健康状态）
		return &consulapi.AgentServiceCheck{
			TTL:                            c.config.HealthCheckTTL,
			DeregisterCriticalServiceAfter: c.config.DeregisterCriticalServiceAfter,
		}

	case HealthCheckTypeNone:
		// 不进行健康检查
		return nil

	default:
		// 默认使用 gRPC 检查
		grpcAddr := fmt.Sprintf("%s:%d", addr, port)
		return &consulapi.AgentServiceCheck{
			GRPC:                           grpcAddr,
			GRPCUseTLS:                     c.config.GRPCUseTLS,
			Interval:                       c.config.HealthCheckInterval,
			Timeout:                        c.config.HealthCheckTimeout,
			DeregisterCriticalServiceAfter: c.config.DeregisterCriticalServiceAfter,
		}
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
