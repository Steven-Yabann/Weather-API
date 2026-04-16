package main

import (
	"encoding/json"
	"log"
	"net/http"

    "github.com/Steven-Yabann/weather-api/config"
    "github.com/Steven-Yabann/weather-api/middleware"
    "github.com/Steven-Yabann/weather-api/weather"
)

func main() {
	cfg := config.Load()
	// rateLimiter := ratelimit.NewRateLimiter()

	weatherClient, err := weather.NewClient(cfg.APIKey, cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to init weather client: %v", err)
	}

	r1 := middleware.NewRateLimiter()

	//	This creates a multiplexer
	//	This is the engine that looks at an incoming URL path
	//	decides which function should handle it
	mux := http.NewServeMux()

	// Route registration
	// This is the handler for the /weather endpoint
	mux.HandleFunc("/weather", weatherClient.HandleGetWeather)

	handler := r1.Middleware(mux)
	log.Printf("Listening on port: %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}