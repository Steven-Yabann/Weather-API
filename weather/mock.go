package weather

import (
	"context"
	"net/http"
)

type MockClient struct {}

func (m *MockClient) GetWeather (ctx context.Context, city string) (*WeatherData, error) {
	return &WeatherData{City: city, Temperature: 25.0, Description: "Always Sunny"}, nil
}

func (m *MockClient) HandleGetWeather (w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from the Mock"))
}