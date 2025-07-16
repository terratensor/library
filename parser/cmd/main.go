package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/terratensor/library/parser/internal/config"
	"github.com/terratensor/library/parser/internal/lib/logger/handlers/slogpretty"
	"github.com/terratensor/library/parser/internal/lib/logger/sl"
	"github.com/terratensor/library/parser/internal/library/entry"
	"github.com/terratensor/library/parser/internal/metadata"
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

	// Инициализация мета-процессора
	metaCfg := metadata.Config{
		LogFilePath: "metadata_errors.log",
		Logger:      logger,
	}

	metaProcessor, err := metadata.NewProcessor(metaCfg)
	if err != nil {
		logger.Error("failed to initialize metadata processor", sl.Err(err))
		os.Exit(1)
	}
	defer metaProcessor.Close()

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

	// Режим только метаданные
	if cfg.MetadataOnly {
		if err := processMetadata(ctx, metaProcessor, prs, cfg, logger); err != nil {
			logger.Error("error processing metadata", sl.Err(err))
			os.Exit(1)
		}
		return
	}
	// Новый код для обработки tar-архивов
	if isTarArchive(cfg.Volume) {
		if err := processTarArchive(ctx, prs, cfg, logger); err != nil {
			logger.Error("error processing tar archive", sl.Err(err))
			os.Exit(1)
		}
	} else {
		// Старый код для обработки обычных файлов
		files, paths, err := findFiles(cfg.Volume)
		if err != nil {
			logger.Error("error reading directory", sl.Err(err))
			os.Exit(1)
		}

		var allTask []*workerpool.Task
		for n, file := range files {
			task := workerpool.NewTask(func(data interface{}) error {
				fileData := data.(struct {
					file os.DirEntry
					path string
					n    int
				})
				fmt.Printf("Processing file %v\n", fileData.file.Name())
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				err = prs.Parse(ctx, fileData.file, filepath.Dir(fileData.path))
				if err != nil {
					logger.Error("error processing file",
						slog.String("filename", fileData.file.Name()),
						sl.Err(err))
					return err
				}
				return nil
			}, struct {
				file os.DirEntry
				path string
				n    int
			}{file, paths[n], n})
			allTask = append(allTask, task)
		}
		defer utils.Duration(utils.Track("Обработка завершена за "))
		pool := workerpool.NewPool(allTask, cfg.Concurrency)
		pool.Run()
	}

	// Сохраняем все модели перед выходом
	// if err := prs.StoreModels(ctx, metaProcessor); err != nil {
	// 	logger.Error("error storing models", sl.Err(err))
	// }

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

// Новые вспомогательные функции
func isTarArchive(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".tar" || ext == ".tar.gz"
}

func processTarArchive(ctx context.Context, prs *parser.Parser, cfg *config.Config, logger *slog.Logger) error {
	file, err := os.Open(cfg.Volume)
	if err != nil {
		return fmt.Errorf("error opening tar file: %v", err)
	}
	defer file.Close()

	return prs.ProcessTar(ctx, file, cfg.Concurrency)
}

// Функция для рекурсивного поиска всех файлов в директории и поддиректориях (кроме исключений)
func findFiles(rootDir string) ([]os.DirEntry, []string, error) {
	var files []os.DirEntry
	var paths []string

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем директории и файлы из исключений (например, .gitignore)
		if !d.IsDir() && d.Name() != ".gitignore" {
			files = append(files, d)
			paths = append(paths, path)
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return files, paths, nil
}

func processMetadata(ctx context.Context, metaProcessor *metadata.Processor, prs *parser.Parser, cfg *config.Config, logger *slog.Logger) error {
	logger.Info("running in metadata-only mode")

	files, paths, err := findFiles(cfg.Volume)
	if err != nil {
		logger.Error("error reading directory", sl.Err(err))
		os.Exit(1)
	}

	for n, file := range files {
		if err := prs.ProcessMetadataOnly(ctx, metaProcessor, file, filepath.Dir(paths[n])); err != nil {
			logger.Error("error processing file metadata",
				slog.String("filename", file.Name()),
				sl.Err(err))
		}
	}

	// Сохраняем отчеты
	if err := metaProcessor.SaveDuplicatesReport("duplicates_report.txt"); err != nil {
		logger.Error("failed to save duplicates report", sl.Err(err))
	}

	// Сохраняем модели в базу
	if err := prs.StoreModels(ctx, metaProcessor); err != nil {
		logger.Error("failed to store models", sl.Err(err))
		os.Exit(1)
	}

	report := metaProcessor.GenerateReport()
	logger.Info("metadata processing completed",
		slog.Int("files_processed", len(report.Entries)),
		slog.Int("authors", len(metaProcessor.GetAuthors())),
		slog.Int("categories", len(metaProcessor.GetCategories())),
		slog.Int("titles", len(metaProcessor.GetTitles())),
		slog.Int("duplicates_found", len(report.Duplicates)))
	return nil
}
