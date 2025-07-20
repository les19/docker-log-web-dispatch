package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	LoggerServiceURL      string
	LoggerAuthHeaderName  string
	LoggerAuthHeaderValue string
	HTTPClientTimeout     time.Duration
	LogTailCount          string
	ListenAll             bool
	ContainerNameFilters  []string
}

const (
	defaultLoggerServiceURL      = "http://127.0.0.1"
	defaultHTTPClientTimeoutSecs = 10 // seconds
	defaultLogTailCount          = "10"
	defaultContainerNameFilters  = "app,application"
)

func LoadConfig() *Config {
	cfg := &Config{
		LoggerServiceURL:      getEnv("LOGGER_SERVICE_URL", defaultLoggerServiceURL),
		LoggerAuthHeaderName:  getEnv("LOGGER_AUTH_HEADER_NAME", "Authorization"),
		LoggerAuthHeaderValue: getEnv("LOGGER_AUTH_HEADER_VALUE", ""),
		HTTPClientTimeout:     time.Duration(getEnvInt("HTTP_CLIENT_TIMEOUT_SECONDS", defaultHTTPClientTimeoutSecs)) * time.Second,
		LogTailCount:          getEnv("LOG_TAIL_COUNT", defaultLogTailCount),
		ListenAll:             getEnvBool("LISTEN_ALL_CONTAINERS", false),
		ContainerNameFilters:  getEnvSlice("CONTAINER_NAME_FILTERS", defaultContainerNameFilters),
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		} else {
			log.Printf("Warning: Invalid integer for environment variable %s, using default %d. Error: %v", key, defaultValue, err)
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		parsedValue, err := strconv.ParseBool(value)
		if err == nil {
			return parsedValue
		}
		log.Printf("Warning: Invalid boolean for environment variable %s, using default %t. Error: %v", key, defaultValue, err)
	}
	return defaultValue
}

func getEnvSlice(key, defaultValue string) []string {
	value := getEnv(key, defaultValue)
	if value == "" {
		return []string{} // Return empty slice if no value or empty string
	}

	// Trim spaces around each element after splitting
	parts := strings.Split(value, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}
