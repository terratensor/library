package metadata

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/terratensor/library/parser/internal/library/book"
	"github.com/terratensor/library/parser/internal/library/entry"
)

type Processor struct {
	duplicates map[string][]string // key: title, value: список путей к файлов
	entries    map[string]book.TitleList
	authors    map[string]entry.Author
	categories map[string]entry.Category
	titles     map[string]entry.Title

	dupMutex    sync.Mutex
	entryMutex  sync.Mutex
	modelsMutex sync.Mutex

	errorLog *os.File
	logger   *slog.Logger
}

type Config struct {
	LogFilePath string
	Logger      *slog.Logger
}

func NewProcessor(cfg Config) (*Processor, error) {
	f, err := os.OpenFile(cfg.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	return &Processor{
		duplicates: make(map[string][]string),
		entries:    make(map[string]book.TitleList),
		authors:    make(map[string]entry.Author),
		categories: make(map[string]entry.Category),
		titles:     make(map[string]entry.Title),
		errorLog:   f,
		logger:     cfg.Logger,
	}, nil
}

func (mp *Processor) Close() error {
	if err := mp.errorLog.Sync(); err != nil {
		return fmt.Errorf("failed to sync error log: %v", err)
	}
	if err := mp.errorLog.Close(); err != nil {
		return fmt.Errorf("failed to close error log: %v", err)
	}
	return nil
}

func (mp *Processor) ProcessFile(path string) error {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)

	// Пропускаем неподдерживаемые форматы
	switch ext {
	case ".docx", ".pdf", ".epub":
		// Продолжаем обработку
	default:
		return nil
	}

	bookName := filename[:len(filename)-len(ext)]
	titleList := book.NewTitleList(bookName)
	if titleList.Title == "" {
		msg := fmt.Sprintf("invalid filename format: %s", filename)
		mp.logError(msg)
		return fmt.Errorf(msg)
	}

	titleList.SourceUUID = uuid.New()
	titleList.Source = filename

	// Обработка дубликатов
	mp.processDuplicates(titleList, path)

	// Обработка моделей (авторы, категории, заголовки)
	mp.processModels(titleList)

	// Сохраняем запись
	mp.entryMutex.Lock()
	mp.entries[path] = *titleList
	mp.entryMutex.Unlock()

	return nil
}

func (mp *Processor) processDuplicates(tl *book.TitleList, path string) {
	mp.dupMutex.Lock()
	defer mp.dupMutex.Unlock()

	if files, exists := mp.duplicates[tl.Title]; exists {
		mp.duplicates[tl.Title] = append(files, path)
		msg := fmt.Sprintf("duplicate title '%s' in files: %v", tl.Title, mp.duplicates[tl.Title])
		mp.logWarning(msg)
	} else {
		// Проверяем, есть ли такой заголовок в уже обработанных файлах
		for p, entry := range mp.entries {
			if entry.Title == tl.Title {
				mp.duplicates[tl.Title] = []string{p, path}
				msg := fmt.Sprintf("duplicate title '%s' found between: %s and %s", tl.Title, p, path)
				mp.logWarning(msg)
				return
			}
		}
		mp.duplicates[tl.Title] = []string{path}
	}
}

func (mp *Processor) logError(msg string) error {
	if _, err := mp.errorLog.WriteString(fmt.Sprintf("[ERROR] %s\n", msg)); err != nil {
		return fmt.Errorf("failed to write to error log: %v", err)
	}
	mp.logger.Error(msg)
	return nil
}

func (mp *Processor) logWarning(msg string) {
	mp.logger.Warn(msg)
	mp.errorLog.WriteString(fmt.Sprintf("[WARN] %s\n", msg))
}

func (mp *Processor) GetAuthors() []entry.Author {
	mp.modelsMutex.Lock()
	defer mp.modelsMutex.Unlock()

	authors := make([]entry.Author, 0, len(mp.authors))
	for _, a := range mp.authors {
		authors = append(authors, a)
	}
	return authors
}

func (mp *Processor) GetCategories() []entry.Category {
	mp.modelsMutex.Lock()
	defer mp.modelsMutex.Unlock()

	categories := make([]entry.Category, 0, len(mp.categories))
	for _, c := range mp.categories {
		categories = append(categories, c)
	}
	return categories
}

func (mp *Processor) GetTitles() []entry.Title {
	mp.modelsMutex.Lock()
	defer mp.modelsMutex.Unlock()

	titles := make([]entry.Title, 0, len(mp.titles))
	for _, t := range mp.titles {
		titles = append(titles, t)
	}
	return titles
}

func (mp *Processor) processModels(tl *book.TitleList) {
	mp.modelsMutex.Lock()
	defer mp.modelsMutex.Unlock()

	// Обработка автора
	if tl.Author != "" {
		if _, exists := mp.authors[tl.Author]; !exists {
			author := entry.NewAuthorFromTitleList(tl)
			mp.authors[tl.Author] = *author
		}
	}

	// Обработка категории
	if tl.Genre != "" {
		if _, exists := mp.categories[tl.Genre]; !exists {
			category := entry.NewCategoryFromTitleList(tl)
			mp.categories[tl.Genre] = *category
		}
	}

	// Обработка заголовка
	if tl.Title != "" {
		if _, exists := mp.titles[tl.Title]; !exists {
			title := entry.NewTitleFromTitleList(tl)
			mp.titles[tl.Title] = *title
		}
	}
}
