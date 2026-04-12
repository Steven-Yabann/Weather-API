package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// object response from server
type WeatherResponse struct {
	City		string 		`json:"city"`
	Temperature	float64 	`json:"temperature_c"`
	Description	string 		`json:"description"`
	Humidity	int			`json:"humidity"`
}

// handler
// w output back to client
// r input from client
func weatherHandler (w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")

	// handle bad query. No city parameter provided
	if city == "" {
		http.Error(w, `{"error":"city query param required"}`, http.StatusBadRequest)
		return
	}

	// hardcoded now; we'll work on later
	resp := WeatherResponse {
		City: 			city,
		Temperature: 	24.5,
		Description: 	"Partly Cloudy",
		Humidity: 		65,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/weather", weatherHandler)
	log.Println("Server running on: 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}