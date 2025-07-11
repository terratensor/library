package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	// Чтение API-ключа из файла (Docker-секрет)
	apiKeyBytes, err := os.ReadFile(os.Getenv("MANTICORE_API_KEY_FILE"))
	if err != nil {
		log.Fatal("Failed to read API key:", err)
	}
	apiKey := string(apiKeyBytes)

	// Адрес Manticore из переменной окружения
	manticoreHost := os.Getenv("MANTICORE_HOST")
	if manticoreHost == "" {
		log.Fatal("MANTICORE_HOST is not set")
	}

	// Настройка прокси
	manticoreURL, err := url.Parse("http://" + manticoreHost)
	if err != nil {
		log.Fatal("Failed to parse Manticore URL:", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(manticoreURL)

	// Обработчик запросов
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != apiKey {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid API key"))
			return
		}
		proxy.ServeHTTP(w, r)
	})

	// Запуск сервера
	listenPort := ":9308" // Внутри Docker-сети
	log.Printf("Starting proxy (Manticore: %s, Listen: %s)", manticoreHost, listenPort)
	log.Fatal(http.ListenAndServe(listenPort, nil))
}
