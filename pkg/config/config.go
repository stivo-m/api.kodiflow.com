package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// Returns an env variable for the given key
// Panics if the variable is not found
func MustGetEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Panicf("Environment variable %s was not found", key)
	}

	return value
}

func GetStringEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}

// GetIntEnv works by casting the obtained env value
// into an integer or revers with the fallback
func GetIntEnv(key string, fallback int) int {
	value := GetStringEnv(key, fmt.Sprintf("%v", fallback))

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return intValue
}

// GetIntEnv works by casting the obtained env value
// into an float or revers with the fallback
func GetFloatEnv(key string, fallback float64) float64 {
	value := GetStringEnv(key, fmt.Sprintf("%v", fallback))

	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}

	return floatVal
}

