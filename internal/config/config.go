package config

import (
	"log"
	"os"
	"strconv"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
}

type Config struct {
	SMTP SMTPConfig
}

// getEnv 读取环境变量，如果不存在则返回默认值
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Using fallback for env var %s", key)
	return fallback
}

// LoadConfig 从环境变量中加载配置
func LoadConfig() Config {
	portStr := getEnv("SMTP_PORT", "587")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Invalid SMTP_PORT: %s. Using default 587.", portStr)
		port = 587
	}

	return Config{
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp-relay.brevo.com"),
			Port:     port,
			Username: getEnv("SMTP_USERNAME", ""), // 没有默认值，必须提供
			Password: getEnv("SMTP_PASSWORD", ""), // 没有默认值，必须提供
			Sender:   getEnv("SMTP_SENDER", ""),   // 没有默认值，必须提供
		},
	}
}