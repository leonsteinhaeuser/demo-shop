package env

import "os"

func StringEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func BytesEnvOrDefault(key string, defaultValue []byte) []byte {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return []byte(value)
}
