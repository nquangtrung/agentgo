package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv(filenames ...string) {
	log.Println("Loading environment variables from .env file...")
	envErr := godotenv.Load(filenames...)
	if envErr != nil {
		log.Fatal(envErr)
		panic(".env file not found")
	}
}

func GetEnvVar(key string) string {
	value := os.Getenv(key)
	return value
}
