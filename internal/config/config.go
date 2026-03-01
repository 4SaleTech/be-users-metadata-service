package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration.
type Config struct {
	DB         DBConfig      // Primary DB: event_sources, metadata_rules, processed_events, failed_events
	UsersDB    UsersDBConfig // Users DB: clas_users (or configured table) for meta_data updates
	RabbitMQ   RabbitMQConfig
	Server     ServerConfig
	Processing ProcessingConfig
	Log        LogConfig
}

// UsersDBConfig holds the second DB connection used to update the users/clas_users table.
type UsersDBConfig struct {
	DSN             string
	TableName       string // e.g. "clas_users"
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// LogConfig holds logging settings (LOG_LEVEL: debug, info, warn, error).
type LogConfig struct {
	Level string
}

// DBConfig holds database settings (from env).
type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RabbitMQConfig holds RabbitMQ connection settings (from env).
// AMQP (port 5672) uses URL or RABBITMQ_* vars. Stream/Super Stream (port 5552) uses Stream* and same user/password.
type RabbitMQConfig struct {
	URL             string
	Prefetch        int
	User            string
	Password        string
	StreamEnabled   bool
	SuperStreamName string
	StreamHost      string
	StreamPort      int
}

// ServerConfig holds graceful shutdown settings.
type ServerConfig struct {
	ShutdownTimeout time.Duration
}

// ProcessingConfig holds consumer and processing settings.
type ProcessingConfig struct {
	Workers   int
	BatchSize int
}

// Load reads configuration from environment variables.
// Values from the OS environment take precedence; if unset, variables are loaded from .env (if present).
//
// Database: DATABASE_URL (primary), DATABASE_USERS_URL (optional, for clas_users). RabbitMQ: RABBITMQ_URL or RABBITMQ_*.
func Load() *Config {
	loadEnvFile(".env")
	return &Config{
		DB:         loadDBConfig(),
		UsersDB:    loadUsersDBConfig(),
		RabbitMQ:   loadRabbitMQConfig(),
		Server:     loadServerConfig(),
		Processing: loadProcessingConfig(),
		Log:        loadLogConfig(),
	}
}

// loadEnvFile loads variables from path into the environment (does not override existing).
func loadEnvFile(path string) {
	_ = godotenv.Load(path)
}

func loadDBConfig() DBConfig {
	return DBConfig{
		DSN:             getEnv("DATABASE_URL", ""),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
	}
}

// loadUsersDBConfig loads config for the DB that holds clas_users. DATABASE_USERS_URL only; if empty, single-DB mode.
func loadUsersDBConfig() UsersDBConfig {
	return UsersDBConfig{
		DSN:             getEnv("DATABASE_USERS_URL", ""),
		TableName:       getEnv("USERS_DB_TABLE_NAME", "clas_users"),
		MaxOpenConns:    getEnvInt("USERS_DB_MAX_OPEN_CONNS", 10),
		MaxIdleConns:    getEnvInt("USERS_DB_MAX_IDLE_CONNS", 3),
		ConnMaxLifetime: getEnvDuration("USERS_DB_CONN_MAX_LIFETIME", 5*time.Minute),
	}
}

func loadRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		URL:             buildRabbitMQURL(),
		Prefetch:        getEnvInt("RABBITMQ_PREFETCH", 10),
		User:            getEnv("RABBITMQ_USER", "guest"),
		Password:        getEnv("RABBITMQ_PASSWORD", "guest"),
		StreamEnabled:   getEnv("RABBITMQ_STREAM_ENABLED", "false") == "true",
		SuperStreamName: getEnv("RABBITMQ_SUPER_STREAM_NAME", ""),
		StreamHost:      getEnv("RABBITMQ_STREAM_HOST", getEnv("RABBITMQ_HOST", "localhost")),
		StreamPort:      getEnvInt("RABBITMQ_STREAM_PORT", 5552),
	}
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
	}
}

func loadProcessingConfig() ProcessingConfig {
	return ProcessingConfig{
		Workers:   getEnvInt("WORKERS", 5),
		BatchSize: getEnvInt("BATCH_SIZE", 100),
	}
}

func loadLogConfig() LogConfig {
	return LogConfig{
		Level: getEnv("LOG_LEVEL", "info"),
	}
}

// buildRabbitMQURL returns AMQP URL: RABBITMQ_URL if set, else built from RABBITMQ_* env vars.
func buildRabbitMQURL() string {
	if v := os.Getenv("RABBITMQ_URL"); v != "" {
		return v
	}
	user := getEnv("RABBITMQ_USER", "guest")
	password := getEnv("RABBITMQ_PASSWORD", "guest")
	host := getEnv("RABBITMQ_HOST", "localhost")
	port := getEnv("RABBITMQ_PORT", "5672")
	vhost := getEnv("RABBITMQ_VHOST", "/")
	if vhost == "/" {
		return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, password, host, port, vhost)
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}
