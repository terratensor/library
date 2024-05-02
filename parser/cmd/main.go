package main

import (
	"context"
	"fmt"
	"github.com/terratensor/library/parser/internal/config"
	"github.com/terratensor/library/parser/internal/errorlog"
	"github.com/terratensor/library/parser/internal/library/entry"
	"github.com/terratensor/library/parser/internal/parser"
	"github.com/terratensor/library/parser/internal/storage/manticore"
	"github.com/terratensor/library/parser/internal/utils"
	"github.com/terratensor/library/parser/internal/workerpool"
	"log"
	"os"
	"os/signal"
)

func main() {
	// Чтение конфиг-файла
	cfg := config.MustLoad()
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	// Инициализация хранилища
	manticoreClient, err := manticore.New(ctx, &cfg.Manticore)
	if err != nil {
		log.Fatalf("error creating manticore client: %v", err)
	}
	storage := entry.New(manticoreClient)

	// Инициализация парсера
	prs := parser.NewParser(cfg, storage)

	// читаем все файлы в директории
	files, err := os.ReadDir(cfg.Volume)
	if err != nil {
		log.Fatal(err)
	}

	// Срез ошибок полученных при обработке файлов
	var errors []string

	var allTask []*workerpool.Task

	// Цикл обработки файлов
	for n, file := range files {
		if file.IsDir() == false {

			// если файл gitignore, то ничего не делаем пропускаем и продолжаем цикл
			if file.Name() == ".gitignore" {
				continue
			}

			// добавляем задание в пул
			task := workerpool.NewTask(func(data interface{}) error {

				fmt.Printf("Task %v processed\n", file.Name())
				// обрабатываем файл
				err := prs.Parse(ctx, n, file, cfg.Volume)
				if err != nil {
					return err
				}
				return nil
			}, file)

			allTask = append(allTask, task)
			errors = append(errors, fmt.Sprintln(err))
		}
	}
	defer utils.Duration(utils.Track("Обработка завершена за "))
	pool := workerpool.NewPool(allTask, cfg.Concurrency)
	pool.Run()

	errorlog.Save(errors)
	log.Println("all files done")
}
