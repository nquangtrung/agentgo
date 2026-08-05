package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatal("Error loading .env file")
		panic(".env file not found")
	}
}

func GetEnvVar(key string) string {
	LoadEnv()
	value := os.Getenv(key)
	return value
}
