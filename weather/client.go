package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"errors"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	apiKey		string
	httpClient	*http.Client
	cache		*redis.Client
	cacheTTL	time.Duration
}

type WeatherData struct {
	City        string  `json:"city"`
    Temperature float64 `json:"temperature_c"`
    Description string  `json:"description"`
    Humidity    int     `json:"humidity_pct"`
    FromCache   bool    `json:"from_cache"`
}

// constructor func
func NewClient (apiKey, redisURL string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)

	if err != nil {
		return nil, fmt.Errorf("Invalid REDIS url: %w", err)
	}

	return &Client{
		apiKey: 	apiKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		cache: 		redis.NewClient(opts),
		cacheTTL: 	10 * time.Minute,
	}, nil
}

type owmResponse struct {
    Main struct {
        Temp     float64 `json:"temp"`
        Humidity int     `json:"humidity"`
    } `json:"main"`
    Weather []struct {
        Description string `json:"description"`
    } `json:"weather"`
    Name    string `json:"name"`
    Cod     int `json:"cod"`    
    Message string `json:"message"`
}

func (c *Client) GetWeather (ctx context.Context, city string) (*WeatherData, error) {
	cacheKey := "weather:" + city

	// check cache first
	// The .Get retreives the memory location of the value. Result() is used to get the value
	cached, err := c.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var data WeatherData
		if err := json.Unmarshal([]byte(cached), &data); err == nil {
			data.FromCache = true	// Mark as cache hit
			return &data, nil
		}
		return nil, fmt.Errorf("Unmarshal error: %v \n", err)
	}

	// Call OpenWeatherMap
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
        city, c.apiKey,
	)

	// Create a http.Request object. No request is sent; configuration is done here
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Building request: %w", err)
	}

	// Request is sent
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Upstream API unavailable: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("OWM status: %d, body city: %s", resp.StatusCode, city)

	// Map API error codes to meaningful errors
	switch resp.StatusCode {
	case http.StatusOK:
		// continue
	case http.StatusNotFound:
		return nil, &CityNotFoundError{City : city}
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("Invalid API Key")
	default:
		return nil, fmt.Errorf("upstream error %d: %s", resp.StatusCode, resp.Body)
	}

	// declare a OWM object to hold the response
	var owm owmResponse

	// NewDecoder() reads the response body in a streaming fashion for memory efficiency
	// .Decode() parses the response body into the owm object
	if err := json.NewDecoder(resp.Body).Decode(&owm); err != nil {
		return nil, fmt.Errorf("Decoding response: %w", err)
	}

	// transformation phase
	// Map the OpenWeatherMap response to our WeatherData
	// We use pointers for efficiency
	data := &WeatherData {
		City:			owm.Name,
		Temperature:	owm.Main.Temp,
		Humidity:		owm.Main.Humidity,
	}

	if len(owm.Weather) > 0 {
		data.Description = owm.Weather[0].Description
	}

	// Store in cache
	// Marshal() converts go struct (data) into a byte slice formatted as JSON
	if b, err := json.Marshal(data); err == nil {
		c.cache.Set(ctx, cacheKey, b, c.cacheTTL)
	}

	return data, nil
}


// Typed error lets handlers distinguish "city not found" from "API down"
type CityNotFoundError struct { City string }

func ( e *CityNotFoundError ) Error() string {
	return fmt.Sprintf("city not found: %s", e.City)
}


