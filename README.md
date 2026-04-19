# Multi-Provider Weather API & System Design Lab

A high-performance, concurrent Go backend service built to demonstrate **Interface Abstraction**, **Resilient System Design**, and **Distributed Caching**.

## Key Architectural Features

* **Strategy Design Pattern**: Implements a `WeatherProvider` interface allowing seamless hot-swapping between **OpenWeatherMap** and **Visual Crossing** without modifying core business logic.
* **Distributed Caching**: Uses **Redis** with provider-specific namespacing (`vc:weather:city`) to minimize external API latency and manage rate-limit costs.
* **Resiliency Layer (Token Bucket)**: Features a custom-built, thread-safe Middleware Rate Limiter using `sync.Mutex` to protect upstream resources from IP-based exhaustion.
* **Dependency Injection**: Decouples HTTP transport logic from data retrieval, enabling easier unit testing and mock provider implementations.

---

## System Workflow

1. **Request Ingress**: Client requests weather via `GET /weather?city=Nairobi`.
2. **Rate Limiting**: Middleware checks the IP's token bucket; if empty, returns `429 Too Many Requests`.
3. **Cache Lookup**: The system queries Redis for a cached version of the data (10-minute TTL).
4. **API Fetch (Strategy)**: On a cache miss, the active `WeatherProvider` executes a context-aware HTTP request.
5. **Data Normalization**: External JSON responses are mapped into a consistent internal `WeatherData` struct, ensuring consistent units and data types regardless of the source.

---

## Technical Stack

| Component | Technology | Purpose |
| :--- | :--- | :--- |
| **Language** | Go (Golang) | High-concurrency backend logic |
| **Cache** | Redis | Latency reduction and API cost management |
| **Concurrency** | Goroutines & Mutexes | Thread-safe rate limiting and background cache cleanup |
| **Env Management** | `godotenv` | Secure handling of API credentials and configurations |

---

## Challenges Overcome

* **JSON Structural Mismatch**: Resolved issues mapping complex nested API arrays (Visual Crossing's `days` array) into flat Go structs.
* **Context Deadlines**: Hardened the system against upstream latency by implementing 10-second `context.WithTimeout` on all external calls.
* **State Management**: Built a background `cleanup()` worker in the rate limiter to prevent memory leaks from stale IP tracking.