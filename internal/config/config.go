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
}

type HttpApiConfig struct {
	Address  string `yaml:"address"`
	TLS      bool   `yaml:"tls"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

func Load(configPath string, logger *slog.Logger) (*Config, error) {
	return commonConfig.Load[Config](configPath, logger)
}
