package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// struct for variables in the .env
type Config struct {
	APIKey		string
	RedisURL	string
	Port		string
}

// core function to load .env values and return the object
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file")
	}

	return &Config{
		APIKey: 	mustGet("OPEN_WEATHER_API_KEY"),
		RedisURL: 	getOrDefault("REDIS_URL", "redis://localhost:6379"),
		Port:		getOrDefault("Port", "8080"),
	}
}

// assistant function to retreive the API key. If none, do not start
func mustGet(key string) string {
	v := os.Getenv(key)

	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}

	return v
}

// assistant function to provide an alternative if value of config struct isnt available in .env
func getOrDefault(key string, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}