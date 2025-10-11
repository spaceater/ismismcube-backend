package config

import (
	"log"
	"os"
	"strconv"
)

var (
	Port             string
	PageViewFile     string
	ExecutedTaskFile string
)

func Init() {
	Port = getEnv("PORT", "2998")
	PageViewFile = getEnv("PAGE_VIEW_FILE", "./resources/page-view.txt")
	ExecutedTaskFile = getEnv("EXECUTED_TASK_FILE", "./resources/executed-task.txt")
	InitWSConfig()
	InitLLMConfig()
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		log.Printf("getEnv: %s: %s", key, value)
		return value
	}
	log.Printf("getEnv: %s: %s", key, defaultValue)
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			log.Printf("getEnv: %s: %s", key, value)
			return intValue
		}
	}
	log.Printf("getEnv: %s: %d", key, defaultValue)
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			log.Printf("getEnv: %s: %s", key, value)
			return floatValue
		}
	}
	log.Printf("getEnv: %s: %f", key, defaultValue)
	return defaultValue
}
