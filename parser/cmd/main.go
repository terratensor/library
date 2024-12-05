package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"

	"github.com/terratensor/library/parser/internal/config"
	"github.com/terratensor/library/parser/internal/lib/logger/handlers/slogpretty"
	"github.com/terratensor/library/parser/internal/lib/logger/sl"
	"github.com/terratensor/library/parser/internal/library/entry"
	"github.com/terratensor/library/parser/internal/parser"
	"github.com/terratensor/library/parser/internal/storage/manticore"
	"github.com/terratensor/library/parser/internal/utils"
	"github.com/terratensor/library/parser/internal/workerpool"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// Чтение конфиг-файла
	cfg := config.MustLoad()

	logger := setupLogger(cfg.Env)
	logger = logger.With(slog.String("env", cfg.Env)) // к каждому сообщению будет добавляться поле с информацией о текущем окружении

	logger.Debug("logger debug mode enabled")
	logger.Debug("initializing manticore client",
		slog.String("index", cfg.Manticore.Index),
		slog.String("host", cfg.Manticore.Host),
		slog.String("port", cfg.Manticore.Port),
	)

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	// Инициализация хранилища
	manticoreClient, err := manticore.New(ctx, &cfg.Manticore)
	if err != nil {
		logger.Error("error creating manticore client", sl.Err(err))
		os.Exit(1)
	}
	storage := entry.New(manticoreClient)

	// Инициализация парсера
	prs := parser.NewParser(cfg, storage)

	// читаем все файлы в директории
	files, err := os.ReadDir(cfg.Volume)
	if err != nil {
		logger.Error("error reading directory", sl.Err(err))
		os.Exit(1)
	}

	// Срез ошибок полученных при обработке файлов
	// var errors []string

	var allTask []*workerpool.Task

	// Цикл обработки файлов
	for n, file := range files {
		if !file.IsDir() {

			// если файл gitignore, то ничего не делаем пропускаем и продолжаем цикл
			if file.Name() == ".gitignore" {
				continue
			}

			// добавляем задание в пул
			task := workerpool.NewTask(func(data interface{}) error {

				fmt.Printf("Task %v processed\n", file.Name())

				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				// обрабатываем файл
				err = prs.Parse(ctx, n, file, cfg.Volume)
				if err != nil {
					logger.Error("error processing file", sl.Err(err))
					return err
				}
				return nil
			}, file)

			allTask = append(allTask, task)
		}
	}
	defer utils.Duration(utils.Track("Обработка завершена за "))
	pool := workerpool.NewPool(allTask, cfg.Concurrency)
	pool.Run()

	// errorlog.Save(errors)
	log.Println("all files done")
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = setupPrettySlog()
	case envDev:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return logger
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
