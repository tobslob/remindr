package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}
}

func GetString(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("environment variable '%s' is not set", key))
	}
	return val
}

func GetInt(key string) int {
	valStr, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("environment variable '%s' is not set", key))
	}
	valInt, err := strconv.Atoi(valStr)
	if err != nil {
		panic(fmt.Sprintf("environment variable '%s' is not a valid integer", key))
	}

	return valInt
}
