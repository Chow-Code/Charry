package cluster

import "time"

// NacosConfig Nacos 配置
type NacosConfig struct {
	// Nacos 服务器配置
	ServerConfigs []ServerConfig `json:"server_configs"`
	// 客户端配置
	ClientConfig ClientConfig `json:"client_config"`
	// 集群配置
	ClusterConfig ClusterConfig `json:"cluster_config"`
}

// ServerConfig Nacos 服务器配置
type ServerConfig struct {
	IpAddr      string `json:"ip_addr"`      // Nacos 服务器 IP
	Port        uint64 `json:"port"`         // Nacos 服务器端口
	ContextPath string `json:"context_path"` // 上下文路径
	Scheme      string `json:"scheme"`       // 协议 http 或 https
}

// ClientConfig Nacos 客户端配置
type ClientConfig struct {
	NamespaceId         string `json:"namespace_id"`            // 命名空间 ID
	TimeoutMs           uint64 `json:"timeout_ms"`              // 超时时间（毫秒）
	NotLoadCacheAtStart bool   `json:"not_load_cache_at_start"` // 不在启动时加载缓存
	LogDir              string `json:"log_dir"`                 // 日志目录
	CacheDir            string `json:"cache_dir"`               // 缓存目录
	LogLevel            string `json:"log_level"`               // 日志级别
	Username            string `json:"username"`                // 用户名
	Password            string `json:"password"`                // 密码
}

// ClusterConfig 集群配置
type ClusterConfig struct {
	DataId          string        `json:"data_id"`           // 配置文件 Data ID
	Group           string        `json:"group"`             // 配置文件分组
	WatchInterval   time.Duration `json:"watch_interval"`    // 监控间隔
	NodeSyncTimeout time.Duration `json:"node_sync_timeout"` // 节点同步超时时间
}

// DefaultNacosConfig 获取默认 Nacos 配置
func DefaultNacosConfig() *NacosConfig {
	return &NacosConfig{
		ServerConfigs: []ServerConfig{
			{
				IpAddr:      "127.0.0.1",
				Port:        8848,
				ContextPath: "/nacos",
				Scheme:      "http",
			},
		},
		ClientConfig: ClientConfig{
			NamespaceId:         "",
			TimeoutMs:           5000,
			NotLoadCacheAtStart: true,
			LogDir:              "/tmp/nacos/log",
			CacheDir:            "/tmp/nacos/cache",
			LogLevel:            "info",
		},
		ClusterConfig: ClusterConfig{
			DataId:          "cluster-nodes",
			Group:           "DEFAULT_GROUP",
			WatchInterval:   time.Second * 5,
			NodeSyncTimeout: time.Second * 30,
		},
	}
}
