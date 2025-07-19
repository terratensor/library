package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

// ErrorResponse структура для JSON-ошибок
type ErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Config содержит конфигурацию прокси
type Config struct {
	APIKey          string
	ManticoreHost   string
	ProxyListenPort string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	DialTimeout     time.Duration
	IdleTimeout     time.Duration
	RequestTimeout  time.Duration
	MaxHeaderBytes  int
}

func main() {
	// Инициализация конфигурации
	cfg := loadConfig()

	// Настройка прокси с таймаутами
	proxy := createReverseProxy(cfg)

	// Настройка HTTP сервера
	server := &http.Server{
		Addr:           cfg.ProxyListenPort,
		Handler:        createHandler(proxy, cfg.APIKey),
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		IdleTimeout:    cfg.IdleTimeout,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}

	// Запуск сервера
	log.Printf("Starting proxy (Manticore: %s, Listen: %s)", cfg.ManticoreHost, cfg.ProxyListenPort)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func loadConfig() Config {
	// Чтение API-ключа из файла (Docker-секрет)
	apiKeyBytes, err := os.ReadFile(os.Getenv("MANTICORE_API_KEY_FILE"))
	if err != nil {
		log.Fatal("Failed to read API key:", err)
	}

	// Проверка обязательных переменных окружения
	manticoreHost := os.Getenv("MANTICORE_HOST")
	if manticoreHost == "" {
		log.Fatal("MANTICORE_HOST is not set")
	}

	return Config{
		APIKey:          string(apiKeyBytes),
		ManticoreHost:   manticoreHost,
		ProxyListenPort: getEnv("PROXY_LISTEN_PORT", ":9308"),
		ReadTimeout:     getEnvDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getEnvDuration("WRITE_TIMEOUT", 10*time.Second),
		DialTimeout:     getEnvDuration("DIAL_TIMEOUT", 5*time.Second),
		IdleTimeout:     getEnvDuration("IDLE_TIMEOUT", 30*time.Second),
		RequestTimeout:  getEnvDuration("REQUEST_TIMEOUT", 15*time.Second),
		MaxHeaderBytes:  getEnvInt("MAX_HEADER_BYTES", 1<<20), // 1MB
	}
}

func createReverseProxy(cfg Config) *httputil.ReverseProxy {
	target, err := url.Parse("http://" + cfg.ManticoreHost)
	if err != nil {
		log.Fatal("Failed to parse Manticore URL:", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       cfg.IdleTimeout,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
		},
		Transport: transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %v", err)
			sendJSONError(w, "Bad Gateway", "Service unavailable", http.StatusBadGateway)
		},
	}
}

func createHandler(proxy *httputil.ReverseProxy, apiKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Логирование входящего запроса
		log.Printf("Incoming request: %s %s", r.Method, r.URL.Path)

		// Устанавливаем заголовки ответа
		w.Header().Set("Content-Type", "application/json")

		// Проверка API ключа
		if r.Header.Get("X-API-Key") != apiKey {
			sendJSONError(w, "Unauthorized", "Invalid API key", http.StatusUnauthorized)
			return
		}

		// Проксирование запроса
		proxy.ServeHTTP(w, r)
	})
}

func sendJSONError(w http.ResponseWriter, errorType, message string, statusCode int) {
	w.WriteHeader(statusCode)
	response := ErrorResponse{}
	response.Error.Type = errorType
	response.Error.Message = message
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// Вспомогательные функции для работы с переменными окружения
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		duration, err := time.ParseDuration(value)
		if err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		var result int
		if _, err := fmt.Sscan(value, &result); err == nil {
			return result
		}
	}
	return defaultValue
}
