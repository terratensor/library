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
	Model string `json:"model"`
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
		vectorizerHost = "http://vectorizer:8081" // дефолтное значение
		log.Printf("Using default vectorizer host: %s", vectorizerHost)
	}

	vectorizerPyHost := os.Getenv("VECTORIZER_PY_HOST")
	if vectorizerPyHost == "" {
		vectorizerPyHost = "http://vectorizer-py:8082"
		log.Printf("Using default python vectorizer host: %s", vectorizerPyHost)
	}

	return Config{
		APIKey:           string(apiKeyBytes),
		ManticoreHost:    manticoreHost,
		VectorizerHost:   vectorizerHost,
		VectorizerPyHost: vectorizerPyHost,
		ProxyListenPort:  getEnv("PROXY_LISTEN_PORT", ":9308"),
		ReadTimeout:      getEnvDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:     getEnvDuration("WRITE_TIMEOUT", 10*time.Second),
		DialTimeout:      getEnvDuration("DIAL_TIMEOUT", 5*time.Second),
		IdleTimeout:      getEnvDuration("IDLE_TIMEOUT", 30*time.Second),
		RequestTimeout:   getEnvDuration("REQUEST_TIMEOUT", 15*time.Second),
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

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
		},
		Transport: transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error to %s: %v", targetURL, err)
			sendJSONError(w, "Bad Gateway", "Service unavailable", http.StatusBadGateway)
		},
	}
}

func createHandler(manticoreProxy, vectorizerProxy, vectorizerPyProxy *httputil.ReverseProxy, apiKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем заголовки ответа
		w.Header().Set("Content-Type", "application/json")

		// Проверка API ключа
		if r.Header.Get("X-API-Key") != apiKey {
			w.WriteHeader(http.StatusUnauthorized)
			response := map[string]interface{}{
				"error": "Invalid API key",
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				log.Printf("Error encoding JSON response: %v", err)
			}
			return
		}

		if r.URL.Path == "/vectorize-py" || strings.EqualFold(r.Header.Get("X-Vectorizer"), "py") {
			handleVectorizationRequest(w, r, vectorizerPyProxy)
			return
		}

		// Определяем тип запроса
		if isVectorizationRequest(r) {
			// Перехватываем запрос для векторизации
			handleVectorizationRequest(w, r, vectorizerProxy)
			return
		}

		// Остальные запросы перенаправляем в Manticore
		manticoreProxy.ServeHTTP(w, r)
	})
}

func isVectorizationRequest(r *http.Request) bool {
	// Вариант 1: По специальному пути
	if r.URL.Path == "/vectorize" {
		return true
	}

	// Вариант 2: По заголовку
	if r.Header.Get("X-Request-Type") == "vectorization" {
		return true
	}

	// Вариант 3: По содержимому запроса
	if r.Method == http.MethodPost {
		// Проверяем первые 100 байт тела на наличие признаков векторизации
		bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 100))
		if err == nil {
			r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(bodyBytes), r.Body))
			return bytes.Contains(bodyBytes, []byte(`"text"`))
		}
	}

	return false
}

func handleVectorizationRequest(w http.ResponseWriter, r *http.Request, vectorizerProxy *httputil.ReverseProxy) {
	// Читаем тело запроса
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		sendJSONError(w, "Bad Request", "Failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	// Декодируем запрос
	var req VectorizationRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		sendJSONError(w, "Bad Request", "Invalid request format", http.StatusBadRequest)
		return
	}

	// Логируем запрос (без текста для безопасности)
	log.Printf("Vectorization request (text length: %d)", len(req.Text))

	// Перенаправляем запрос в vectorizer-server
	vectorizerProxy.ServeHTTP(w, r)
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
