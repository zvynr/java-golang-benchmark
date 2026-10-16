package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port              string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBMaxConns        int32
	DBMinConns        int32
	DBMaxConnLifetime string
	DBMaxConnIdleTime string
	DBConnectTimeout  string
}

func LoadConfig() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		DBHost:            getEnv("DB_HOST", "postgres"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "benchmark"),
		DBPassword:        getEnv("DB_PASSWORD", "benchmark"),
		DBName:            getEnv("DB_NAME", "benchmark"),
		DBMaxConns:        int32(getEnvAsInt("DB_MAX_CONNS", 20)), // Matches Java HikariCP maximum-pool-size: 20
		DBMinConns:        int32(getEnvAsInt("DB_MIN_CONNS", 5)),  // Matches Java HikariCP minimum-idle: 5
		DBMaxConnLifetime: getEnv("DB_MAX_CONN_LIFETIME", "30m"),  // Matches Java HikariCP max-lifetime: 30m
		DBMaxConnIdleTime: getEnv("DB_MAX_CONN_IDLE_TIME", "10m"),  // Matches Java HikariCP idle-timeout: 10m
		DBConnectTimeout:  getEnv("DB_CONNECT_TIMEOUT", "30s"), // Matches Java HikariCP connection-timeout: 30s
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}
