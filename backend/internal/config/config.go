package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App   AppConfig
	Auth  AuthConfig
	MySQL MySQLConfig
}

type AppConfig struct {
	Name            string
	Env             string
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type AuthConfig struct {
	TokenSecret string
	TokenTTL    time.Duration
}

func (a AppConfig) Address() string {
	return fmt.Sprintf("%s:%s", a.Host, a.Port)
}

type MySQLConfig struct {
	Enabled         bool
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	Charset         string
	Collation       string
	ParseTime       bool
	Loc             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func Load() (Config, error) {
	loadDotEnv()

	cfg := Config{
		App: AppConfig{
			Name:            getEnv("APP_NAME", "SIAKAD Backend"),
			Env:             getEnv("APP_ENV", "development"),
			Host:            getEnv("APP_HOST", "0.0.0.0"),
			Port:            getEnv("APP_PORT", "8080"),
			ReadTimeout:     getDurationEnv("APP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDurationEnv("APP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     getDurationEnv("APP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getDurationEnv("APP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Auth: AuthConfig{
			TokenSecret: getEnv("AUTH_TOKEN_SECRET", "dev-secret-change-me"),
			TokenTTL:    getDurationEnv("AUTH_TOKEN_TTL", 24*time.Hour),
		},
		MySQL: MySQLConfig{
			Enabled:         getBoolEnv("MYSQL_ENABLED", true),
			Host:            getEnv("MYSQL_HOST", "127.0.0.1"),
			Port:            getIntEnv("MYSQL_PORT", 3306),
			User:            getEnv("MYSQL_USER", "root"),
			Password:        getEnv("MYSQL_PASSWORD", ""),
			Database:        getEnv("MYSQL_DATABASE", "siakad_db"),
			Charset:         getEnv("MYSQL_CHARSET", "utf8mb4"),
			Collation:       getEnv("MYSQL_COLLATION", "utf8mb4_unicode_ci"),
			ParseTime:       getBoolEnv("MYSQL_PARSE_TIME", true),
			Loc:             getEnv("MYSQL_LOC", "Local"),
			MaxOpenConns:    getIntEnv("MYSQL_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("MYSQL_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getDurationEnv("MYSQL_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getDurationEnv("MYSQL_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.App.Port == "" {
		return fmt.Errorf("APP_PORT is required")
	}

	if c.Auth.TokenSecret == "" {
		return fmt.Errorf("AUTH_TOKEN_SECRET is required")
	}

	if !c.MySQL.Enabled {
		return nil
	}

	if c.MySQL.Host == "" {
		return fmt.Errorf("MYSQL_HOST is required when MYSQL_ENABLED=true")
	}

	if c.MySQL.User == "" {
		return fmt.Errorf("MYSQL_USER is required when MYSQL_ENABLED=true")
	}

	if c.MySQL.Database == "" {
		return fmt.Errorf("MYSQL_DATABASE is required when MYSQL_ENABLED=true")
	}

	return nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getIntEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getBoolEnv(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func loadDotEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		// Only set if not already set by environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
