package config

import (
	"os"
	"strconv"
)

// EnvArgs 环境变量参数
// 应用启动后第一个加载，所有环境变量统一从这里获取
type EnvArgs struct {
	// 应用配置
	AppId   uint16
	AppType string
	AppHost string
	AppPort int

	// Consul 配置
	ConsulAddress                 string
	ConsulDatacenter              string
	ConsulHealthCheckType         string
	ConsulHealthCheckPath         string
	ConsulHealthCheckInterval     string
	ConsulHealthCheckTimeout      string
	ConsulDeregisterCriticalAfter string
	ConsulHealthCheckTTL          string
	ConsulGRPCUseTLS              bool
}

// LoadEnvArgs 从环境变量加载所有配置参数
// 这是启动后第一个调用的方法
func LoadEnvArgs() *EnvArgs {
	return &EnvArgs{
		// 应用配置
		AppId:   uint16(getEnvAsInt("APP_ID", 1)),
		AppType: getEnv("APP_TYPE", "app-service"),
		AppHost: getEnv("APP_HOST", "0.0.0.0"),
		AppPort: getEnvAsInt("APP_PORT", 50051),

		// Consul 配置
		ConsulAddress:                 getEnv("CONSUL_ADDRESS", "localhost:8500"),
		ConsulDatacenter:              getEnv("CONSUL_DATACENTER", "dc1"),
		ConsulHealthCheckType:         getEnv("CONSUL_HEALTH_CHECK_TYPE", "tcp"),
		ConsulHealthCheckPath:         getEnv("CONSUL_HEALTH_CHECK_PATH", ""),
		ConsulHealthCheckInterval:     getEnv("CONSUL_HEALTH_CHECK_INTERVAL", "10s"),
		ConsulHealthCheckTimeout:      getEnv("CONSUL_HEALTH_CHECK_TIMEOUT", "5s"),
		ConsulDeregisterCriticalAfter: getEnv("CONSUL_DEREGISTER_CRITICAL_SERVICE_AFTER", "30s"),
		ConsulHealthCheckTTL:          getEnv("CONSUL_HEALTH_CHECK_TTL", "30s"),
		ConsulGRPCUseTLS:              getEnv("CONSUL_GRPC_USE_TLS", "false") == "true",
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为 int，如果不存在或转换失败则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
