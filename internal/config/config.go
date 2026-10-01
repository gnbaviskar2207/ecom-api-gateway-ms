package config

import (
	"log/slog"

	commonConfig "github.com/gnbaviskar2207/ecom-common/pkg/config"
)

type GatewayConfig struct {
	Name string `yaml:"name"`
}

type Config struct {
	GatewayConfig GatewayConfig `yaml:"gateway"`
	HttpApiConfig HttpApiConfig `yaml:"http_api"`
	ProductConfig ProductConfig `yaml:"product_config"`
}

type ProductConfig struct {
	Address string `yaml:"address"`
}

type HttpApiConfig struct {
	Address            string `yaml:"address"`
	TLS                bool   `yaml:"tls"`
	CertFile           string `yaml:"cert_file"`
	KeyFile            string `yaml:"key_file"`
	ShitDownTimeoutSec int    `yaml:"shutdown_timeout_sec"`
}

func Load(configPath string, logger *slog.Logger) (*Config, error) {
	return commonConfig.Load[Config](configPath, logger)
}
