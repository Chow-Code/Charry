package config

import (
	"os"
	"strconv"
)

// Config 应用程序主配置结构
type Config struct {
	App AppConfig `mapstructure:"app" yaml:"app" json:"app"`
}

// AppConfig 应用配置
type AppConfig struct {
	Id          uint16         `mapstructure:"id" yaml:"id" json:"id"`
	Type        string         `mapstructure:"type" yaml:"type" json:"type"`
	Environment string         `mapstructure:"environment" yaml:"environment" json:"environment"` // dev, test, prod
	Addr        Addr           `mapstructure:"addr" yaml:"addr" json:"addr"`
	Metadata    map[string]any `mapstructure:"metadata" yaml:"metadata" json:"metadata"`
}

// Addr 地址配置
type Addr struct {
	Host string `mapstructure:"host" yaml:"host" json:"host"`
	Port int    `mapstructure:"port" yaml:"port" json:"port"`
}

// LoadAddrFromEnv 从环境变量加载 Addr 配置
// 环境变量：
//   - APP_HOST: 监听主机（可选，默认 "0.0.0.0"）
//   - APP_PORT: 监听端口（可选，默认 "50051"）
func LoadAddrFromEnv() Addr {
	host := getEnv("APP_HOST", "0.0.0.0")
	port := getEnvAsInt("APP_PORT", 50051)

	return Addr{
		Host: host,
		Port: port,
	}
}

// LoadIdFromEnv 从环境变量加载 ID
// 环境变量：
//   - APP_ID: 应用实例 ID（可选，默认 "1"）
func LoadIdFromEnv() uint16 {
	return uint16(getEnvAsInt("APP_ID", 1))
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
