package main

import (
	"encoding/json"
	"log"
	"net/http"
	"context"
	"errors"

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

	mux := http.NewServeMux()
	mux.HandleFunc("/weather", func (w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("city")

		if city == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "city is required"})
			return
		}

		data, err := weatherClient.GetWeather(context.Background(), city)
		if err != nil {
			var notFound *weather.CityNotFoundError
			if errors.As(err, &notFound) {
				writeJSON(w, http.StatusNotFound, map[string]string{"error":err.Error()})
				return
			}
			
			// Upstream is down
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error":"weather service unavailable"})
			return
		}

		writeJSON(w, http.StatusOK, data)
	})

	handler := r1.Middleware(mux)
	log.Printf("Listening on port: %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}