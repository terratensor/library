package entry

import (
	"context"
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/abadojack/whatlanggo"
	"github.com/google/uuid"
)

// PrepareParagraphs срез подготовленных параграфов книги
type PrepareParagraphs []Entry

type Entry struct {
	ID         *int64    `json:"id,omitempty"`
	SourceUUID uuid.UUID `json:"source_uuid"`
	Source     string    `json:"source"`
	Genre      string    `json:"genre"`
	Author     string    `json:"author"`
	BookName   string    `json:"title"`
	Content    string    `json:"content"`
	Language   string    `json:"language"` // "ru", "en", "de" и т.д.
	Chunk      int       `json:"chunk"`
	CharCount  int       `json:"char_count"`  // Реальное количество символов
	WordCount  int       `json:"word_count"`  // Количество слов
	OCRQuality float32   `json:"ocr_quality"` // 0.0 - 1.0 (1.0 - идеальное качество)
	Datetime   int64     `json:"datetime"`
	CreatedAt  int64     `json:"created_at"`
	UpdatedAt  int64     `json:"updated_at"`
}

type StorageInterface interface {
	// Bulk index operations Post/bulk
	Bulk(ctx context.Context, entries *[]Entry) error
}

type Entries struct {
	store StorageInterface
}

func New(store StorageInterface) *Entries {
	return &Entries{
		store: store,
	}
}

func (e Entries) Bulk(ctx context.Context, entries []Entry) error {
	const op = "entry.Entries.Bulk"

	select {
	case <-ctx.Done():
		return fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	err := e.store.Bulk(ctx, &entries)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (e *Entry) DetectLanguage() {
	info := whatlanggo.Detect(e.Content)
	e.Language = info.Lang.Iso6391() // "ru", "en" и т.д.
}

func (e *Entry) CalculateCharCount() {
	e.CharCount = utf8.RuneCountInString(e.Content)
}

func (e *Entry) CalculateWordCount() {
	inWord := false
	count := 0

	for _, r := range e.Content {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			if !inWord {
				count++
				inWord = true
			}
		} else {
			inWord = false
		}
	}

	e.WordCount = count
}

func (e *Entry) CalculateOCRQuality() {
	strangeChars := 0
	totalChars := 0

	for _, r := range e.Content {
		if r == ' ' || r == '\n' || r == '\t' {
			continue
		}
		totalChars++

		// Проверяем на "странные" символы
		if r > unicode.MaxASCII && !unicode.Is(unicode.Cyrillic, r) && !unicode.Is(unicode.Latin, r) {
			strangeChars++
		}
	}

	if totalChars == 0 {
		e.OCRQuality = 1.0
		return
	}

	quality := 1.0 - float32(strangeChars)/float32(totalChars)
	// Гарантируем диапазон 0.0-1.0
	e.OCRQuality = max(0.0, min(1.0, quality))
}
