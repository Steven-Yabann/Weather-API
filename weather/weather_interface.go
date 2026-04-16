type WeatherProvider interface {
	GetWeather(ctx context.Context, city string) (*WeatherData, error)
}