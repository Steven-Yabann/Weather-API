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

	var weatherProvider weather.WeatherProvider

	client, err := weather.NewVisualCrossingClient(cfg.APIKey, cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to init weather client: %v", err)
	}

	weatherProvider = client

	r1 := middleware.NewRateLimiter()

	//	This creates a multiplexer
	//	This is the engine that looks at an incoming URL path
	//	decides which function should handle it
	mux := http.NewServeMux()

	// Route registration
	// This is the handler for the /weather endpoint
	mux.HandleFunc("/weather", makeWeatherHandler(weatherProvider))

	handler := r1.Middleware(mux)
	log.Printf("Listening on port: %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}

func makeWeatherHandler (p weather.WeatherProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// city := path.Base(r.URL.Path)

		city := r.URL.Query().Get("city")

		if city == ""{
			writeJSON(
				w, 
				http.StatusNotFound, 
				map[string]string {"error":"city is required"},
			)
			return
		}

		data, err := p.GetWeather(r.Context(), city)
		if err != nil {
			writeJSON(
				w,
				http.StatusInternalServerError,
				map[string]string {
					"error":err.Error(),
				},
			)
			return
		}

		writeJSON(w, http.StatusOK, data)

	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}