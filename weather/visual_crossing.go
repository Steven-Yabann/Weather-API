package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

//	Visual Crossing client structure
type VisualCrossingClient struct {
	apiKey		string
	httpClient	*http.Client
	cache		*redis.Client
	cacheTTL	time.Duration
}

// 	vcResponse struct
type vcResponse struct {
	Address		string `json:"address"`
	CurrentConditions struct {
		Temp		float64 `json:"temp"`
		Humidity	float64 `json:"humidity"`
		Condition	string `json:"conditions"`
	} `json:"currentConditions"`
}

//	Visual Crossing client constructor
func NewVisualCrossingClient(apiKey, redisURL string) (*VisualCrossingClient, error) {
	opts, err := redis.ParseURL(redisURL)	// get the redis url connection

	if err != nil {
		return nil, fmt.Errorf("Invalud REDIS url: %w", err)
	}

	return &VisualCrossingClient{
		apiKey: 		apiKey,
		httpClient: 	&http.Client{Timeout: 5 * time.Second},
		cache: 			redis.NewClient(opts),
		cacheTTL: 		10 * time.Minute,
	}, nil
}

//	Align with the weather interface
func (c *VisualCrossingClient) GetWeather(ctx context.Context, city string) (*WeatherData, error) {
	// 1. Check cache 
	// 2. Call Visual Crossing URL
	// 3. Create a http.Request object
	// 4. Send request
	// 5. Map API error codes
    // 6. Parse their specific JSON structure
    // 7. Convert it into the standard *WeatherData

	cacheKey := "vc:weather:" + city

	// Check cache
	cached, err := c.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var data WeatherData
		if err := json.Unmarshal([]byte(cached), &data); err == nil {
			data.FromCache = true
			return &data, nil
		}
	}

	log.Printf("Fetching weather for %s...", city)

	// Call VC URL
	url := fmt.Sprintf(
		"https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/%s?key=%s&unitGroup=metric",
		city, c.apiKey,
	)

	// Create a http.Request object
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Building request: %w", err)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Upstream API unavailable: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("status: %d, body city: %v", resp.StatusCode, resp.Body)

	// Map API error codes
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

	// vcResponse object
	var vcObject vcResponse

	// Parse the response to the vcObject
	if err := json.NewDecoder(resp.Body).Decode(&vcObject); err != nil {
		return nil, fmt.Errorf("Decoding response: %w", err)
	}

	// Map vcResponse object
	data := &WeatherData {
		City: 			vcObject.Address,
		Temperature: 	vcObject.CurrentConditions.Temp,
		Humidity: 		int(vcObject.CurrentConditions.Humidity),
		Description: 	vcObject.CurrentConditions.Condition,
	}

	// Store in Cache
	if b, err := json.Marshal(data); err == nil {
		c.cache.Set(ctx, cacheKey, b, c.cacheTTL)
	}

	return data, nil
}