package env

import (
	"os"
	"strings"
)

func MapEnvOrDefault(key string, defaultValue map[string]string) map[string]string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return parseMap(value)
}

func parseMap(value string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return result
}
