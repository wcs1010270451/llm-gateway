package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env                  string
	HTTPAddr             string
	DatabaseURL          string
	AuthTokenSecret      string
	ProviderKeySecret    string
	FrontendOrigin       string
	GinTrustedProxies    []string
	LogBodyDir           string
	LogBodyRetentionDays int
}

func Load() Config {
	loadLocalEnv()

	authTokenSecret := getenv("AUTH_TOKEN_SECRET", "change-me-in-development")
	return Config{
		Env:               getenv("APP_ENV", "development"),
		HTTPAddr:          getenv("HTTP_ADDR", "127.0.0.1:3212"),
		DatabaseURL:       getenv("DATABASE_URL", "postgres://llm_gateway:llm_gateway@127.0.0.1:55432/llm_gateway?sslmode=disable"),
		AuthTokenSecret:   authTokenSecret,
		ProviderKeySecret: getenv("PROVIDER_KEY_SECRET", authTokenSecret),
		FrontendOrigin:    getenv("FRONTEND_ORIGIN", "http://127.0.0.1:3213,http://localhost:3213"),
		GinTrustedProxies: splitCSV(
			getenv("GIN_TRUSTED_PROXIES", "127.0.0.1"),
		),
		LogBodyDir:           getenv("LOG_BODY_DIR", "./tmp/request-log-bodies"),
		LogBodyRetentionDays: getenvInt("LOG_BODY_RETENTION_DAYS", 7),
	}
}

func loadLocalEnv() {
	for _, path := range []string{"configs/.env", "../configs/.env", ".env"} {
		if loadEnvFile(path) {
			return
		}
	}
}

func loadEnvFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, value)
	}
	return true
}

func (c Config) Validate() error {
	if c.Env != "production" {
		return nil
	}

	if c.AuthTokenSecret == "" || c.AuthTokenSecret == "change-me-in-development" || strings.HasPrefix(c.AuthTokenSecret, "replace-with-") || len(c.AuthTokenSecret) < 32 {
		return fmt.Errorf("AUTH_TOKEN_SECRET must be changed and at least 32 characters in production")
	}
	if c.ProviderKeySecret == "" || c.ProviderKeySecret == "change-me-provider-key-secret" || strings.HasPrefix(c.ProviderKeySecret, "replace-with-") || len(c.ProviderKeySecret) < 32 {
		return fmt.Errorf("PROVIDER_KEY_SECRET must be changed and at least 32 characters in production")
	}
	if strings.Contains(c.FrontendOrigin, "*") {
		return fmt.Errorf("FRONTEND_ORIGIN must not contain wildcard origins in production")
	}
	if strings.HasPrefix(c.HTTPAddr, "127.0.0.1:") || strings.HasPrefix(c.HTTPAddr, "localhost:") {
		return fmt.Errorf("HTTP_ADDR must listen on a container-accessible address in production")
	}

	return nil
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}
