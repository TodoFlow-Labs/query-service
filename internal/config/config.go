package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	HTTPAddr       string `mapstructure:"http-addr"`
	NATSURL        string `mapstructure:"nats-url"`
	DatabaseURL    string `mapstructure:"database-url"`
	BleveIndexPath string `mapstructure:"bleve-index-path"`
	LogLevel       string `mapstructure:"log-level"`
	MetricsAddr    string `mapstructure:"metrics-addr"`
}

func Load() (*Config, error) {
	// Optional config file
	pflag.String("config", "config.yaml", "Path to config file")

	// Flags
	pflag.String("http-addr", "", "HTTP listen address")
	pflag.String("nats-url", "", "NATS JetStream server URL")
	pflag.String("database-url", "", "Database connection URL")
	pflag.String("bleve-index-path", "", "Path to Bleve index directory")
	pflag.String("log-level", "info", "Log verbosity (debug|info|warn|error)")
	pflag.String("metrics-addr", ":9090", "Metrics listen address")
	pflag.Parse()

	// Bind flags
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, err
	}

	// Support env vars
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Load from file if present
	viper.SetConfigFile(viper.GetString("config"))
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("No config file found, using flags/env: %v\n", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Validate
	if cfg.HTTPAddr == "" {
		return nil, fmt.Errorf("http-addr must be set")
	}
	if cfg.NATSURL == "" {
		return nil, fmt.Errorf("nats-url must be set")
	}
	if cfg.BleveIndexPath == "" {
		return nil, fmt.Errorf("bleve-index-path must be set")
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("database-url must be set")
	}

	return &cfg, nil
}
