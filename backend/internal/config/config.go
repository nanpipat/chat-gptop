package config

import "os"

type Config struct {
	OpenAIKey        string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	StoragePath      string
	BackendPort      string
	GitEncryptionKey string
}

func Load() *Config {
	return &Config{
		OpenAIKey:        getEnv("OPENAI_API_KEY", ""),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "postgres"),
		DBName:           getEnv("DB_NAME", "ragdb"),
		StoragePath:      getEnv("STORAGE_PATH", "./storage"),
		BackendPort:      getEnv("BACKEND_PORT", "8080"),
		GitEncryptionKey: getEnv("GIT_ENCRYPTION_KEY", ""),
	}
}

func (c *Config) DatabaseURL() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=disable"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
