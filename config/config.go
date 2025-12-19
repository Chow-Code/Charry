package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

var (
	// globalConfig 全局配置（私有）
	globalConfig *Config
)

// Config 应用程序主配置结构
type Config struct {
	App          AppConfig    `json:"app"`
	Consul       ConsulConfig `json:"consul"`
	AppConfigKey string       `json:"-"` // Consul KV 配置键（不序列化）
}

// ConsulConfig Consul 配置
type ConsulConfig struct {
	Address                        string `json:"address"`
	Datacenter                     string `json:"datacenter"`
	HealthCheckType                string `json:"health_check_type"`
	HealthCheckPath                string `json:"health_check_path"`
	HealthCheckInterval            string `json:"health_check_interval"`
	HealthCheckTimeout             string `json:"health_check_timeout"`
	DeregisterCriticalServiceAfter string `json:"deregister_critical_service_after"`
	HealthCheckTTL                 string `json:"health_check_ttl"`
	GRPCUseTLS                     bool   `json:"grpc_use_tls"`
}

// AppConfig 应用配置
type AppConfig struct {
	Id          uint16         `json:"id"`
	Type        string         `json:"type"`
	Environment string         `json:"environment"` // dev, test, prod
	Addr        Addr           `json:"addr"`
	Metadata    map[string]any `json:"metadata"`
}

// Addr 地址配置
type Addr struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Init 初始化全局配置
// 从默认配置文件加载，然后应用环境变量
func Init(env *EnvArgs) error {
	// 从默认配置文件加载
	cfg, err := LoadFromFile("default.config.json")
	if err != nil {
		return fmt.Errorf("加载默认配置失败: %w", err)
	}

	// 应用环境变量（直接覆写）
	cfg.App.Id = env.AppId
	cfg.App.Addr.Host = env.AppHost
	cfg.App.Addr.Port = env.AppPort
	cfg.AppConfigKey = env.AppConfigKey
	cfg.Consul.Address = env.ConsulAddress
	cfg.Consul.Datacenter = env.ConsulDatacenter

	// 保存到全局配置
	globalConfig = cfg

	return nil
}

// Get 获取全局配置的副本
// 返回值拷贝，防止外部修改全局配置
func Get() Config {
	if globalConfig == nil {
		return Config{}
	}
	return *globalConfig
}

// getPtr 获取全局配置的指针（内部使用）
// 只在 config 模块内部使用
func getPtr() *Config {
	return globalConfig
}

// mergeFromMap 从 map 合并配置到结构体
// 只处理 JSON 中实际存在的字段
func mergeFromMap(structValue reflect.Value, dataMap map[string]interface{}) error {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		// 获取 JSON 标签名
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// 检查 map 中是否有这个字段
		value, exists := dataMap[jsonTag]
		if !exists || value == nil {
			continue
		}

		// 根据字段类型处理
		if err := setFieldValue(field, value); err != nil {
			return fmt.Errorf("设置字段 %s 失败: %w", fieldType.Name, err)
		}
	}

	return nil
}

// setFieldValue 设置字段值
func setFieldValue(field reflect.Value, value interface{}) error {
	if !field.CanSet() {
		return nil
	}

	valueReflect := reflect.ValueOf(value)

	switch field.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if num, ok := value.(float64); ok {
			field.SetInt(int64(num))
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if num, ok := value.(float64); ok {
			field.SetUint(uint64(num))
		}

	case reflect.Bool:
		if b, ok := value.(bool); ok {
			field.SetBool(b)
		}

	case reflect.Struct:
		// 嵌套结构体
		if subMap, ok := value.(map[string]interface{}); ok {
			return mergeFromMap(field, subMap)
		}

	case reflect.Map:
		// Map 类型
		if mapValue, ok := value.(map[string]interface{}); ok {
			if field.IsNil() {
				field.Set(reflect.MakeMap(field.Type()))
			}
			for k, v := range mapValue {
				field.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
			}
		}

	case reflect.Slice:
		// Slice 类型
		if sliceValue, ok := value.([]interface{}); ok {
			newSlice := reflect.MakeSlice(field.Type(), len(sliceValue), len(sliceValue))
			for i, item := range sliceValue {
				newSlice.Index(i).Set(reflect.ValueOf(item))
			}
			field.Set(newSlice)
		}

	default:
		// 尝试直接设置
		if valueReflect.Type().AssignableTo(field.Type()) {
			field.Set(valueReflect)
		}
	}

	return nil
}

// MergeFromJSON 从 JSON 字符串合并配置到全局配置
// 只解析 JSON 中存在的字段并合并
func MergeFromJSON(jsonStr string) error {
	if jsonStr == "" {
		return nil
	}

	cfg := getPtr()
	if cfg == nil {
		return fmt.Errorf("配置未初始化")
	}

	// 解析 JSON 到 map
	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonMap); err != nil {
		return fmt.Errorf("解析配置 JSON 失败: %w", err)
	}

	// 使用反射合并 JSON 数据
	configValue := reflect.ValueOf(cfg).Elem()
	return mergeFromMap(configValue, jsonMap)
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
