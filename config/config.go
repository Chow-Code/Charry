package config

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
