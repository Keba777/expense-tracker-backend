package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	CORS     CORSConfig
	RateLimit RateLimitConfig
}

type AppConfig struct {
	Name string
	Env  string
	Port int
}

type DBConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecret        string
	RefreshSecret       string
	AccessExpiryMinutes int
	RefreshExpiryDays   int
}

type CORSConfig struct {
	Origins []string
}

type RateLimitConfig struct {
	Max           int
	ExpirySeconds int
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	_ = viper.ReadInConfig()

	cfg := &Config{
		App: AppConfig{
			Name: viper.GetString("APP_NAME"),
			Env:  viper.GetString("APP_ENV"),
			Port: viper.GetInt("APP_PORT"),
		},
		DB: DBConfig{
			Host:            viper.GetString("DB_HOST"),
			Port:            viper.GetInt("DB_PORT"),
			User:            viper.GetString("DB_USER"),
			Password:        viper.GetString("DB_PASSWORD"),
			Name:            viper.GetString("DB_NAME"),
			SSLMode:         viper.GetString("DB_SSLMODE"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: time.Duration(viper.GetInt("DB_CONN_MAX_LIFETIME")) * time.Second,
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetInt("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			AccessSecret:        viper.GetString("JWT_ACCESS_SECRET"),
			RefreshSecret:       viper.GetString("JWT_REFRESH_SECRET"),
			AccessExpiryMinutes: viper.GetInt("JWT_ACCESS_EXPIRY_MINUTES"),
			RefreshExpiryDays:   viper.GetInt("JWT_REFRESH_EXPIRY_DAYS"),
		},
		CORS: CORSConfig{
			Origins: strings.Split(viper.GetString("CORS_ORIGINS"), ","),
		},
		RateLimit: RateLimitConfig{
			Max:           viper.GetInt("RATE_LIMIT_MAX"),
			ExpirySeconds: viper.GetInt("RATE_LIMIT_EXPIRY_SECONDS"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if c.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	if c.DB.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	return nil
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func (c *Config) IsDev() bool {
	return c.App.Env == "development"
}

func (c *Config) IsProd() bool {
	return c.App.Env == "production"
}
