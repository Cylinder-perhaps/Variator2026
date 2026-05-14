package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config описывает полную конфигурацию приложения.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Auth     AuthConfig     `yaml:"auth"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// ServerConfig описывает параметры HTTP-сервера.
type ServerConfig struct {
	Port            int           `yaml:"port" env:"PORT" env-default:"9090"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"15s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"15s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env-default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"10s"`
}

// DatabaseConfig описывает параметры подключения к PostgreSQL.
type DatabaseConfig struct {
	Host            string        `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port            int           `yaml:"port" env:"DB_PORT" env-default:"5432"`
	Name            string        `yaml:"name" env:"DB_NAME" env-default:"variator"`
	User            string        `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Password        string        `yaml:"password" env:"DB_PASS" env-default:"postgres"`
	SSLMode         string        `yaml:"sslmode" env-default:"disable"`
	MaxOpenConns    int           `yaml:"max_open_conns" env-default:"25"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env-default:"10"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env-default:"5m"`
}

// DSN возвращает строку подключения к PostgreSQL.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// RedisConfig описывает параметры подключения к Redis.
type RedisConfig struct {
	Addr         string        `yaml:"addr" env:"REDIS_ADDR" env-default:"localhost:6379"`
	Password     string        `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
	DB           int           `yaml:"db" env-default:"0"`
	PoolSize     int           `yaml:"pool_size" env-default:"20"`
	MinIdleConns int           `yaml:"min_idle_conns" env-default:"5"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"3s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"3s"`
}

// JWTConfig описывает параметры JWT-токенов.
type JWTConfig struct {
	Secret     string        `yaml:"secret" env:"JWT_SECRET" env-default:"your-secret-key-change-this-in-production"`
	AccessTTL  time.Duration `yaml:"access_ttl" env-default:"15m"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" env-default:"720h"`
}

// AuthConfig описывает параметры аутентификации.
type AuthConfig struct {
	BcryptCost       int           `yaml:"bcrypt_cost" env-default:"12"`
	MaxLoginAttempts int           `yaml:"max_login_attempts" env-default:"10"`
	LoginWindow      time.Duration `yaml:"login_window" env-default:"15m"`
}

// LoggingConfig описывает параметры логирования.
type LoggingConfig struct {
	Level  string `yaml:"level" env:"LOG_LEVEL" env-default:"debug"`
	Format string `yaml:"format" env-default:"text"`
}

// MustLoad загружает конфигурацию из переменных окружения.
// При ошибке — вызывает panic.
func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(fmt.Sprintf("failed to read config from env: %v", err))
	}

	return &cfg
}