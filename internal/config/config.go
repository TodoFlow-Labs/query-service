package config

import (
    "fmt"
    "strings"

    "github.com/spf13/pflag"
    "github.com/spf13/viper"
)

// Config holds query-service settings.
type Config struct {
    HTTPAddr       string // e.g. ":8082"
    BleveIndexPath string // e.g. "./index.bleve"
    LogLevel       string // e.g. "debug","info"
	DbUrl 	   string // e.g. "postgresql://root@localhost:26257/defaultdb?sslmode=disable"
}

// Load parses flags / env and returns a Config or error.
func Load() (*Config, error) {
    // Define flags
    pflag.String("http-addr", ":8082", "HTTP listen address")
    pflag.String("bleve-index-path", "./index.bleve", "Path to Bleve index")
    pflag.String("log-level", "info", "Log verbosity (debug|info|warn|error)")
	pflag.String("db-url", "postgresql://root@localhost:26257/defaultdb?sslmode=disable", "Database connection URL")
    pflag.Parse()

    // Bind to viper
    if err := viper.BindPFlags(pflag.CommandLine); err != nil {
        return nil, err
    }
    viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
    viper.AutomaticEnv() // allow BLEVE_INDEX_PATH, HTTP_ADDR, LOG_LEVEL

    // Construct
    cfg := &Config{
        HTTPAddr:       viper.GetString("http-addr"),
        BleveIndexPath: viper.GetString("bleve-index-path"),
        LogLevel:       viper.GetString("log-level"),
		DbUrl:          viper.GetString("db-url"),
    }

    // Validate
    if cfg.HTTPAddr == "" {
        return nil, fmt.Errorf("http-addr must be set")
    }
    if cfg.BleveIndexPath == "" {
        return nil, fmt.Errorf("bleve-index-path must be set")
    }
    if cfg.DbUrl == "" {
        return nil, fmt.Errorf("db-url must be set")
    }
    return cfg, nil
}
