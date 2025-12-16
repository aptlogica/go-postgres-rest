package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
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

// type AppDatabaseConfig struct {
// 	Host         string `mapstructure:"host"`
// 	Port         int    `mapstructure:"port"`
// 	User         string `mapstructure:"user"`
// 	Password     string `mapstructure:"password"`
// 	Name         string `mapstructure:"name"`
// 	SSLMode      string `mapstructure:"ssl_mode"`
// 	MaxOpenConns int    `mapstructure:"max_open_conns"`
// 	MaxIdleConns int    `mapstructure:"max_idle_conns"`
// }

// type UserDatabaseConfig struct {
// 	Host         string `mapstructure:"host"`
// 	Port         int    `mapstructure:"port"`
// 	User         string `mapstructure:"user"`
// 	Password     string `mapstructure:"password"`
// 	Name         string `mapstructure:"name"`
// 	SSLMode      string `mapstructure:"ssl_mode"`
// 	MaxOpenConns int    `mapstructure:"max_open_conns"`
// 	MaxIdleConns int    `mapstructure:"max_idle_conns"`
// }

type AuthConfig struct {
	JWTSecret     string `mapstructure:"jwt_secret"`
	TokenExpiry   int    `mapstructure:"token_expiry"`
	RefreshExpiry int    `mapstructure:"refresh_expiry"`
}

type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	URL      string `mapstructure:"url"`
	Password string `mapstructure:"password"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("auth.token_expiry", 3600)
	viper.SetDefault("auth.refresh_expiry", 86400)
	viper.SetDefault("redis.enabled", false)
	viper.SetDefault("redis.url", "redis://localhost:6379")

	viper.AutomaticEnv()

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
