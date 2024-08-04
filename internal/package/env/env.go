package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// String returns a string that is populated by the environment variable or given a default value.
func String(key, defaultValue string, description string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		value = defaultValue
	}
	return value
}

// Int returns an int  that is populated by the environment variable or given a default value.
func Int(key string, defaultValue int, description string) int {
	var value int
	valueString, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	value, err := strconv.Atoi(valueString)
	if err != nil {
		fmt.Printf("environment variable %s is of invalid type. Value: %s.\n", key, valueString)
		return defaultValue
	}
	return value
}

// Bool returns a bool  that is populated by the environment variable or given a default value.
func Bool(key string, defaultValue bool, description string) bool {
	valueString, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	var value bool
	switch valueString {
	case "1", "t", "T", "true", "TRUE", "True":
		value = true
	case "0", "f", "F", "false", "FALSE", "False":
		value = false
	}

	return value
}

func Duration(key string, defaultValue time.Duration, description string) time.Duration {
	valueString, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	duration, err := time.ParseDuration(valueString)
	if err != nil {
		fmt.Printf("Environment variable %s is of invalid type. Value: %s.\n", key, valueString)
		return defaultValue
	}
	return duration
}
