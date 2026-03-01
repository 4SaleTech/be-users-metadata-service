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
	DB         DBConfig
	RabbitMQ   RabbitMQConfig
	Server     ServerConfig
	Processing ProcessingConfig
	Log        LogConfig
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

// ServerConfig holds HTTP/graceful shutdown settings.
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
// Database: use DATABASE_URL, or build from DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME.
// RabbitMQ: use RABBITMQ_URL, or build from RABBITMQ_USER, RABBITMQ_PASSWORD, RABBITMQ_HOST, RABBITMQ_PORT, RABBITMQ_VHOST.
func Load() *Config {
	loadEnvFile(".env")
	return &Config{
		DB:         loadDBConfig(),
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
		DSN:             buildDBDSN(),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
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

// buildDBDSN returns MySQL DSN: DATABASE_URL if set, else built from DB_* and DATABASE (db name). Reads from OS env first.
func buildDBDSN() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	user := getEnv("DB_USER", "user")
	password := getEnv("DB_PASSWORD", "password")
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "3306")
	dbName := getEnv("DATABASE", "be_users_metadata_service")
	charset := getEnv("DB_CHARSET", "utf8mb4")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True",
		user, password, host, port, dbName, charset,
	)
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
