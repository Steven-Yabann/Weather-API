package weather

import (
	"context"
	"net/http"
)

type WeatherProvider interface {
	GetWeather(ctx context.Context, city string) (*WeatherData, error)

}