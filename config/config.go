package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

// Config 应用程序主配置结构
type Config struct {
	App          AppConfig    `mapstructure:"app" yaml:"app" json:"app"`
	Consul       ConsulConfig `mapstructure:"consul" yaml:"consul" json:"consul"`
	AppConfigKey string       `mapstructure:"-" yaml:"-" json:"-"` // Consul KV 配置键（不序列化）
}

// ConsulConfig Consul 配置
type ConsulConfig struct {
	Address                        string `mapstructure:"address" yaml:"address" json:"address"`
	Datacenter                     string `mapstructure:"datacenter" yaml:"datacenter" json:"datacenter"`
	HealthCheckType                string `mapstructure:"health_check_type" yaml:"health_check_type" json:"health_check_type"`
	HealthCheckPath                string `mapstructure:"health_check_path" yaml:"health_check_path" json:"health_check_path"`
	HealthCheckInterval            string `mapstructure:"health_check_interval" yaml:"health_check_interval" json:"health_check_interval"`
	HealthCheckTimeout             string `mapstructure:"health_check_timeout" yaml:"health_check_timeout" json:"health_check_timeout"`
	DeregisterCriticalServiceAfter string `mapstructure:"deregister_critical_service_after" yaml:"deregister_critical_service_after" json:"deregister_critical_service_after"`
	HealthCheckTTL                 string `mapstructure:"health_check_ttl" yaml:"health_check_ttl" json:"health_check_ttl"`
	GRPCUseTLS                     bool   `mapstructure:"grpc_use_tls" yaml:"grpc_use_tls" json:"grpc_use_tls"`
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

// NewConfigFromEnv 从 EnvArgs 创建完整的 Config
func NewConfigFromEnv(env *EnvArgs) *Config {
	return &Config{
		App: AppConfig{
			Id: env.AppId,
			Addr: Addr{
				Host: env.AppHost,
				Port: env.AppPort,
			},
			Metadata: make(map[string]any),
		},
		Consul: ConsulConfig{
			Address:                        env.ConsulAddress,
			Datacenter:                     env.ConsulDatacenter,
			HealthCheckType:                "tcp",    // 默认值
			HealthCheckInterval:            "10s",   // 默认值
			HealthCheckTimeout:             "5s",    // 默认值
			DeregisterCriticalServiceAfter: "30s",   // 默认值
			HealthCheckTTL:                 "30s",   // 默认值
			GRPCUseTLS:                     false,   // 默认值
		},
		AppConfigKey: env.AppConfigKey,
	}
}

// LoadAddrFromEnv 从 EnvArgs 加载 Addr 配置
func LoadAddrFromEnv(env *EnvArgs) Addr {
	return Addr{
		Host: env.AppHost,
		Port: env.AppPort,
	}
}

// LoadIdFromEnv 从 EnvArgs 加载 ID
func LoadIdFromEnv(env *EnvArgs) uint16 {
	return env.AppId
}

// MergeConfig 合并两个 Config 对象
// 将 config2 中不为空的值合并到 config1 中，返回 config1 的引用
// 用于后续从 Consul 读取配置并覆盖本地配置
func MergeConfig(config1, config2 *Config) *Config {
	if config1 == nil {
		return config2
	}
	if config2 == nil {
		return config1
	}

	// 直接修改 config1 的 App 配置
	config1.App = mergeAppConfig(&config1.App, &config2.App)

	// 合并 Consul 配置
	config1.Consul = mergeConsulConfig(&config1.Consul, &config2.Consul)

	return config1
}

// mergeConsulConfig 合并两个 ConsulConfig
func mergeConsulConfig(consul1, consul2 *ConsulConfig) ConsulConfig {
	result := *consul1

	if consul2.Address != "" {
		result.Address = consul2.Address
	}
	if consul2.Datacenter != "" {
		result.Datacenter = consul2.Datacenter
	}
	if consul2.HealthCheckType != "" {
		result.HealthCheckType = consul2.HealthCheckType
	}
	if consul2.HealthCheckPath != "" {
		result.HealthCheckPath = consul2.HealthCheckPath
	}
	if consul2.HealthCheckInterval != "" {
		result.HealthCheckInterval = consul2.HealthCheckInterval
	}
	if consul2.HealthCheckTimeout != "" {
		result.HealthCheckTimeout = consul2.HealthCheckTimeout
	}
	if consul2.DeregisterCriticalServiceAfter != "" {
		result.DeregisterCriticalServiceAfter = consul2.DeregisterCriticalServiceAfter
	}
	if consul2.HealthCheckTTL != "" {
		result.HealthCheckTTL = consul2.HealthCheckTTL
	}
	// GRPCUseTLS 是 bool，需要特殊处理
	if consul2.GRPCUseTLS {
		result.GRPCUseTLS = consul2.GRPCUseTLS
	}

	return result
}

// mergeAppConfig 合并两个 AppConfig
func mergeAppConfig(app1, app2 *AppConfig) AppConfig {
	result := *app1

	// 合并基本字段（如果 app2 的值不为零值）
	if app2.Id != 0 {
		result.Id = app2.Id
	}
	if app2.Type != "" {
		result.Type = app2.Type
	}
	if app2.Environment != "" {
		result.Environment = app2.Environment
	}

	// 合并 Addr
	result.Addr = mergeAddr(&app1.Addr, &app2.Addr)

	// 合并 Metadata
	if len(app2.Metadata) > 0 {
		if result.Metadata == nil {
			result.Metadata = make(map[string]any)
		}
		for k, v := range app2.Metadata {
			// 只合并不为 nil 的值
			if v != nil && !isZeroValue(v) {
				result.Metadata[k] = v
			}
		}
	}

	return result
}

// mergeAddr 合并两个 Addr
func mergeAddr(addr1, addr2 *Addr) Addr {
	result := *addr1

	if addr2.Host != "" {
		result.Host = addr2.Host
	}
	if addr2.Port != 0 {
		result.Port = addr2.Port
	}

	return result
}

// isZeroValue 判断值是否为零值
func isZeroValue(v interface{}) bool {
	if v == nil {
		return true
	}
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return val.Len() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return val.IsNil()
	}
	return false
}

// ReadFromJSON 从 JSON 字符串读取配置
// 解析 JSON 并合并到当前 Config 对象
func (c *Config) ReadFromJSON(jsonStr string) error {
	if jsonStr == "" {
		return nil
	}

	// 创建临时 Config 用于解析 JSON
	tempConfig := &Config{}

	if err := json.Unmarshal([]byte(jsonStr), tempConfig); err != nil {
		return fmt.Errorf("解析配置 JSON 失败: %w", err)
	}

	// 合并到当前配置
	MergeConfig(c, tempConfig)

	return nil
}

// ToJSON 将配置转换为 JSON 字符串
func (c *Config) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化配置失败: %w", err)
	}
	return string(data), nil
}

// LoadFromFile 从 JSON 文件加载配置
func LoadFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return cfg, nil
}

// ApplyEnvArgs 应用环境变量到配置
// 只覆盖环境变量中设置的值
func (c *Config) ApplyEnvArgs(env *EnvArgs) {
	// 应用配置
	if env.AppId != 0 {
		c.App.Id = env.AppId
	}
	if env.AppHost != "" {
		c.App.Addr.Host = env.AppHost
	}
	if env.AppPort != 0 {
		c.App.Addr.Port = env.AppPort
	}
	if env.AppConfigKey != "" {
		c.AppConfigKey = env.AppConfigKey
	}

	// Consul 配置
	if env.ConsulAddress != "" {
		c.Consul.Address = env.ConsulAddress
	}
	if env.ConsulDatacenter != "" {
		c.Consul.Datacenter = env.ConsulDatacenter
	}
}
