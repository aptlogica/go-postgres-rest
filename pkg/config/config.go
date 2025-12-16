package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
}

type DatabaseConfig struct {
	// ─── Common ─────────────────────────────
	Type         string `mapstructure:"type"` // "sql" or "nosql"
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	DatabaseName string `mapstructure:"database_name"`

	// ─── SQL (GORM) Specific ────────────────
	Driver          string        `mapstructure:"driver"`   // e.g., "postgres", "mysql", "sqlite"
	SSLMode         string        `mapstructure:"ssl_mode"` // e.g., "disable", "require"
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`

	// ─── NoSQL (go-nosql / Mongo) Specific ─
	AuthSource string        `mapstructure:"auth_source"` // default: "admin"
	ReplicaSet string        `mapstructure:"replica_set"` // optional
	URI        string        `mapstructure:"uri"`         // override full URI if needed
	UseSRV     bool          `mapstructure:"use_srv"`     // if using SRV format (e.g., Atlas)
	Timeout    time.Duration `mapstructure:"timeout"`
}

// type DatabaseConfig struct {
// 	Type string `json:"type" yaml:"type"` // e.g. "postgres", "mysql", "sqlite", "mssql", "oracle", "mongodb"
// 	DSN  string `json:"dsn"  yaml:"dsn"`  // complete DSN connection string
// }

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
