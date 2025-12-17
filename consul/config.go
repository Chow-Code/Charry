package consul

import "os"

// HealthCheckType 健康检查类型
type HealthCheckType string

const (
	HealthCheckTypeHTTP HealthCheckType = "http"
	HealthCheckTypeGRPC HealthCheckType = "grpc"
	HealthCheckTypeTCP  HealthCheckType = "tcp"
	HealthCheckTypeTTL  HealthCheckType = "ttl"
	HealthCheckTypeNone HealthCheckType = "none"
)

// Config Consul 配置
type Config struct {
	// Consul 服务器地址，格式：host:port，如 "localhost:8500"
	Address string

	// 数据中心名称，默认为 "dc1"
	Datacenter string

	// 健康检查类型，默认为 "tcp"
	// 可选值：http, grpc, tcp, ttl, none
	HealthCheckType HealthCheckType

	// 健康检查路径（HTTP 或 gRPC 时使用）
	// HTTP 默认：/health
	// gRPC 默认：（空，使用标准 gRPC 健康检查协议）
	HealthCheckPath string

	// 服务注册的健康检查间隔，默认为 "10s"
	HealthCheckInterval string

	// 健康检查超时时间，默认为 "5s"
	HealthCheckTimeout string

	// 注销关键时间，服务多久后被注销，默认为 "30s"
	DeregisterCriticalServiceAfter string

	// TTL 健康检查的间隔（仅当 HealthCheckType 为 ttl 时使用），默认为 "30s"
	HealthCheckTTL string

	// gRPC 健康检查是否使用 TLS，默认为 false
	GRPCUseTLS bool
}

// NewConfigFromEnv 从环境变量创建 Consul 配置
// 环境变量：
//   - CONSUL_ADDRESS: Consul 服务器地址（必需），如 "192.168.30.230:8500"
//   - CONSUL_DATACENTER: 数据中心名称（可选，默认 "dc1"）
//   - CONSUL_HEALTH_CHECK_TYPE: 健康检查类型（可选，默认 "tcp"）
//     可选值：tcp, http, grpc, ttl, none
//   - CONSUL_HEALTH_CHECK_PATH: 健康检查路径（可选）
//     HTTP 默认：/health
//     gRPC 默认：（空，使用标准健康检查服务）
//   - CONSUL_HEALTH_CHECK_INTERVAL: 健康检查间隔（可选，默认 "10s"）
//   - CONSUL_HEALTH_CHECK_TIMEOUT: 健康检查超时（可选，默认 "5s"）
//   - CONSUL_DEREGISTER_CRITICAL_SERVICE_AFTER: 注销关键时间（可选，默认 "30s"）
//   - CONSUL_HEALTH_CHECK_TTL: TTL 健康检查间隔（可选，默认 "30s"）
//   - CONSUL_GRPC_USE_TLS: gRPC 是否使用 TLS（可选，默认 "false"）
func NewConfigFromEnv() *Config {
	cfg := &Config{
		Address:                        getEnv("CONSUL_ADDRESS", "localhost:8500"),
		Datacenter:                     getEnv("CONSUL_DATACENTER", "dc1"),
		HealthCheckType:                HealthCheckType(getEnv("CONSUL_HEALTH_CHECK_TYPE", "tcp")),
		HealthCheckPath:                getEnv("CONSUL_HEALTH_CHECK_PATH", ""),
		HealthCheckInterval:            getEnv("CONSUL_HEALTH_CHECK_INTERVAL", "10s"),
		HealthCheckTimeout:             getEnv("CONSUL_HEALTH_CHECK_TIMEOUT", "5s"),
		DeregisterCriticalServiceAfter: getEnv("CONSUL_DEREGISTER_CRITICAL_SERVICE_AFTER", "30s"),
		HealthCheckTTL:                 getEnv("CONSUL_HEALTH_CHECK_TTL", "30s"),
		GRPCUseTLS:                     getEnv("CONSUL_GRPC_USE_TLS", "false") == "true",
	}
	return cfg
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
