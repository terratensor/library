package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

// ErrorResponse структура для JSON-ошибок
type ErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// ManticoreErrorResponse формат ошибки, который понимает Manticoresearch клиент
type ManticoreErrorResponse struct {
	Error string `json:"error"`
}

// VectorizationRequest структура для запроса векторизации
type VectorizationRequest struct {
	Text  string `json:"text"`
	Model string `json:"model"` // "glove" или "e5-small"
}

// VectorizationResponse структура для ответа векторизации
type VectorizationResponse struct {
	OriginalText string    `json:"original_text"`
	CleanedText  string    `json:"cleaned_text"`
	Vector       []float32 `json:"vector"`
	Method       string    `json:"method"`
	Dimensions   int       `json:"dimensions"`
	Error        string    `json:"error,omitempty"`
}

// Config содержит конфигурацию прокси
type Config struct {
	APIKey           string
	ManticoreHost    string
	VectorizerHost   string
	VectorizerPyHost string
	ProxyListenPort  string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	DialTimeout      time.Duration
	IdleTimeout      time.Duration
	RequestTimeout   time.Duration
	MaxHeaderBytes   int
}

func main() {
	// Инициализация конфигурации
	cfg := loadConfig()

	// Настройка прокси для Manticore
	manticoreProxy := createReverseProxy("http://"+cfg.ManticoreHost, cfg)

	// Настройка прокси для Vectorizer
	vectorizerProxy := createReverseProxy(cfg.VectorizerHost, cfg)
	// Настройка прокси для Vectorizer_py
	vectorizerPyProxy := createReverseProxy(cfg.VectorizerPyHost, cfg)

	// Настройка HTTP сервера
	server := &http.Server{
		Addr:           cfg.ProxyListenPort,
		Handler:        createHandler(manticoreProxy, vectorizerProxy, vectorizerPyProxy, cfg.APIKey),
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		IdleTimeout:    cfg.IdleTimeout,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}

	// Ждем готовности Manticore
	manticoreAddr := strings.TrimPrefix(cfg.ManticoreHost, "http://")
	log.Printf("Waiting for Manticore at %s...", manticoreAddr)
	if err := waitForService(manticoreAddr, 30*time.Second); err != nil {
		log.Fatalf("Manticore is not available: %v", err)
	}

	// Запуск сервера
	log.Printf("Starting proxy (Manticore: %s, Vectorizer: %s, VectorizerPy: %s, Listen: %s)",
		cfg.ManticoreHost, cfg.VectorizerHost, cfg.VectorizerPyHost, cfg.ProxyListenPort)
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

	vectorizerHost := os.Getenv("VECTORIZER_HOST")
	if vectorizerHost == "" {
		vectorizerHost = "http://vectorizer:8081"
		log.Printf("Using default vectorizer host: %s", vectorizerHost)
	}

	vectorizerPyHost := os.Getenv("VECTORIZER_PY_HOST")
	if vectorizerPyHost == "" {
		vectorizerPyHost = "http://vectorizer-py:8082"
		log.Printf("Using default vectorizer-py host: %s", vectorizerPyHost)
	}

	return Config{
		APIKey:           string(apiKeyBytes),
		ManticoreHost:    manticoreHost,
		VectorizerHost:   vectorizerHost,
		VectorizerPyHost: vectorizerPyHost,
		ProxyListenPort:  getEnv("PROXY_LISTEN_PORT", ":9308"),
		ReadTimeout:      getEnvDuration("READ_TIMEOUT", 60*time.Second),  // синхронизировано
		WriteTimeout:     getEnvDuration("WRITE_TIMEOUT", 60*time.Second), // синхронизировано
		DialTimeout:      getEnvDuration("DIAL_TIMEOUT", 5*time.Second),   // как network_timeout
		IdleTimeout:      getEnvDuration("IDLE_TIMEOUT", 270*time.Second), // 4.5m < client_timeout=5m
		RequestTimeout:   getEnvDuration("REQUEST_TIMEOUT", 60*time.Second),
		MaxHeaderBytes:   getEnvInt("MAX_HEADER_BYTES", 1<<20), // 1MB
	}
}

func createReverseProxy(targetURL string, cfg Config) *httputil.ReverseProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL %s: %v", targetURL, err)
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

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
		},
		Transport: transport,
	}

	// ErrorHandler с retry
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		msg := err.Error()
		if strings.Contains(msg, "broken pipe") ||
			strings.Contains(msg, "connection reset by peer") ||
			strings.Contains(msg, "i/o timeout") {

			log.Printf("Retrying request to %s after error: %v", targetURL, err)

			// Копируем тело, если оно было
			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}

			retryReq := r.Clone(r.Context())
			if bodyBytes != nil {
				retryReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}

			proxy.ServeHTTP(w, retryReq)
			return
		}

		log.Printf("Proxy error to %s: %v", targetURL, err)
		sendJSONError(w, "Bad Gateway", "Service unavailable", http.StatusBadGateway)
	}

	return proxy
}

func createHandler(manticoreProxy, vectorizerProxy, vectorizerPyProxy *httputil.ReverseProxy, apiKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем заголовки ответа
		w.Header().Set("Content-Type", "application/json")

		// Проверка API ключа
		if r.Header.Get("X-API-Key") != apiKey {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid API key"})
			return
		}

		if r.URL.Path == "/vectorize" {
			handleVectorizationRequest(w, r, vectorizerProxy, vectorizerPyProxy)
			return
		}

		// Остальные запросы перенаправляем в Manticore
		manticoreProxy.ServeHTTP(w, r)
	})
}

func handleVectorizationRequest(w http.ResponseWriter, r *http.Request, vectorizerProxy, vectorizerPyProxy *httputil.ReverseProxy) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		sendJSONError(w, "Bad Request", "Failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var req VectorizationRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		sendJSONError(w, "Bad Request", "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = "glove" // дефолт
	}

	log.Printf("Vectorization request (model: %s, text length: %d)", req.Model, len(req.Text))

	// Выбор прокси
	switch strings.ToLower(req.Model) {
	case "glove":
		vectorizerProxy.ServeHTTP(w, r)
	case "e5-small":
		vectorizerPyProxy.ServeHTTP(w, r)
	default:
		sendJSONError(w, "Bad Request", fmt.Sprintf("Unknown model: %s", req.Model), http.StatusBadRequest)
	}
}

func sendJSONError(w http.ResponseWriter, errorType, message string, statusCode int) {
	w.WriteHeader(statusCode)
	response := ErrorResponse{}
	response.Error.Type = errorType
	response.Error.Message = message
	json.NewEncoder(w).Encode(response)
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

// Функция ожидания доступности сервиса
func waitForService(addr string, timeout time.Duration) error {
	start := time.Now()
	for {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout after %s", timeout)
		}
		time.Sleep(1 * time.Second)
	}
}
