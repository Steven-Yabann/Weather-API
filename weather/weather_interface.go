package weather

import (
	"context"
)

type WeatherProvider interface {
	GetWeather(ctx context.Context, city string) (*WeatherData, error)

}