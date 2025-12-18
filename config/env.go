package config

import (
	"encoding/json"
	"os"
	"strconv"
)

// EnvArgs 环境变量参数
// 应用启动后第一个加载，所有环境变量统一从这里获取
// 只包含最核心的配置，其他配置从 Consul KV 读取
type EnvArgs struct {
	// 应用配置
	AppId        uint16
	AppHost      string
	AppPort      int
	AppConfigKey string // Consul KV 配置键

	// Consul 配置（只保留必需的连接信息）
	ConsulAddress    string
	ConsulDatacenter string
}

// LoadEnvArgs 从环境变量加载所有配置参数
// 这是启动后第一个调用的方法
func LoadEnvArgs() *EnvArgs {
	return &EnvArgs{
		// 应用配置
		AppId:        uint16(getEnvAsInt("APP_ID", 1)),
		AppHost:      getEnv("APP_HOST", "0.0.0.0"),
		AppPort:      getEnvAsInt("APP_PORT", 50051),
		AppConfigKey: getEnv("APP_CONFIG_KEY", ""),

		// Consul 配置（只保留必需的连接信息）
		ConsulAddress:    getEnv("CONSUL_ADDRESS", "localhost:8500"),
		ConsulDatacenter: getEnv("CONSUL_DATACENTER", "dc1"),
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

// ToJSON 将 EnvArgs 转换为美化的 JSON 字符串
func (e *EnvArgs) ToJSON() string {
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}
