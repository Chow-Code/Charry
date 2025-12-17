package consul

// HealthCheckType 健康检查类型
type HealthCheckType string

const (
	HealthCheckTypeHTTP HealthCheckType = "http"
	HealthCheckTypeGRPC HealthCheckType = "grpc"
	HealthCheckTypeTCP  HealthCheckType = "tcp"
	HealthCheckTypeTTL  HealthCheckType = "ttl"
	HealthCheckTypeNone HealthCheckType = "none"
)

// Config Consul 配置（内部使用）
type Config struct {
	Address                        string
	Datacenter                     string
	HealthCheckType                HealthCheckType
	HealthCheckPath                string
	HealthCheckInterval            string
	HealthCheckTimeout             string
	DeregisterCriticalServiceAfter string
	HealthCheckTTL                 string
	GRPCUseTLS                     bool
}
