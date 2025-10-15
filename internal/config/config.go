package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
)

func Init() {
	var configFilePath string
	flag.StringVar(&configFilePath, "config", "", "Configuration file path (JSON format)")
	flag.Parse()
	var configData map[string]interface{}
	if configFilePath != "" {
		var err error
		configData, err = LoadConfigFromFile(configFilePath)
		if err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}
		log.Printf("Configuration loaded from: %s", configFilePath)
	} else {
		configData = nil
		log.Printf("No config file provided, using default values")
	}
	InitServerConfig(configData)
	InitWSConfig(configData)
	InitLLMConfig(configData)
}

func LoadConfigFromFile(filepath string) (map[string]interface{}, error) {
	if filepath == "" {
		return nil, fmt.Errorf("need to specify the configuration file path in Command Line Args")
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var configData map[string]interface{}
	if err := json.Unmarshal(data, &configData); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return configData, nil
}

func getConfigString(jsonKey string, configData map[string]interface{}, defaultValue string) string {
	var result string
	if configData == nil {
		result = defaultValue
	} else {
		value, exists := configData[jsonKey]
		if !exists {
			result = defaultValue
		} else if strValue, ok := value.(string); ok {
			if strValue == "" {
				result = defaultValue
			} else {
				result = strValue
			}
		} else {
			result = defaultValue
		}
	}
	log.Printf(`[Config] %s = "%s"`, jsonKey, result)
	return result
}

func getConfigInt(jsonKey string, configData map[string]interface{}, defaultValue int) int {
	var result int
	if configData == nil {
		result = defaultValue
	} else {
		value, exists := configData[jsonKey]
		if !exists {
			result = defaultValue
		} else if intValue, ok := value.(int); ok {
			result = intValue
		} else if floatValue, ok := value.(float64); ok {
			result = int(floatValue)
		} else {
			result = defaultValue
		}
	}
  log.Printf(`[Config] %s = %d`, jsonKey, result)
	return result
}

func getConfigFloat(jsonKey string, configData map[string]interface{}, defaultValue float64) float64 {
	var result float64
	if configData == nil {
		result = defaultValue
	} else {
		value, exists := configData[jsonKey]
		if !exists {
			result = defaultValue
		} else if floatValue, ok := value.(float64); ok {
			result = floatValue
		} else if intValue, ok := value.(int); ok {
			result = float64(intValue)
		} else {
			result = defaultValue
		}
	}
	log.Printf(`[Config] %s = %f`, jsonKey, result)
	return result
}

func getConfigStringSlice(jsonKey string, configData map[string]interface{}, defaultValue []string) []string {
	var result []string
	if configData == nil {
		result = defaultValue
	} else {
		value, exists := configData[jsonKey]
		if !exists {
			result = defaultValue
		} else if arrayValue, ok := value.([]interface{}); ok {
			temp := make([]string, 0, len(arrayValue))
			for _, item := range arrayValue {
				if strItem, ok := item.(string); ok {
					temp = append(temp, strItem)
				}
			}
			if len(temp) == 0 {
				result = defaultValue
			} else {
				result = temp
			}
		} else {
			result = defaultValue
		}
	}
	log.Printf(`[Config] %s = %v`, jsonKey, result)
	return result
}

func getJSONTag(structType interface{}, fieldName string) string {
	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	field, found := t.FieldByName(fieldName)
	if !found {
		return ""
	}
	tag := field.Tag.Get("json")
	if tag == "" {
		return ""
	}
	parts := strings.Split(tag, ",")
	return parts[0]
}
