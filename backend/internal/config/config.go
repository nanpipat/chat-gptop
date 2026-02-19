package config

import "os"

type Config struct {
	OpenAIKey   string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBUrl       string
	StoragePath string
	// R2 Configuration
	R2Endpoint        string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Bucket          string

	BackendPort      string
	GitEncryptionKey string
}

func Load() *Config {
	return &Config{
		OpenAIKey:   getEnv("OPENAI_API_KEY", ""),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "ragdb"),
		DBUrl:       getEnv("DATABASE_URL", ""),
		StoragePath: getEnv("STORAGE_PATH", "./storage"),

		R2Endpoint:        getEnv("R2_ENDPOINT", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2Bucket:          getEnv("R2_BUCKET", ""),

		BackendPort:      getEnv("BACKEND_PORT", "8080"),
		GitEncryptionKey: getEnv("GIT_ENCRYPTION_KEY", ""),
	}
}

func (c *Config) DatabaseURL() string {
	if c.DBUrl != "" {
		return c.DBUrl
	}
	return "postgres://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=disable"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
