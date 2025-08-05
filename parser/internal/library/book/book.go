package book

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type TitleList struct {
	SourceUUID uuid.UUID
	Source     string
	Genre      string
	Author     string
	Title      string
	Folder     string // Добавлено новое поле
}

// NewTitleList создает новый TitleList из полного пути файла
func NewTitleList(filePath string, genresMap, foldersMap map[string]string) *TitleList {
	// Извлекаем имя файла и папки
	filename := filepath.Base(filePath)
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
	folder := filepath.Base(filepath.Dir(filePath))

	// Применяем маппинг папок
	if mapped, ok := foldersMap[folder]; ok {
		folder = mapped
	}

	tl := &TitleList{
		Folder: folder,
	}

	// Улучшенное регулярное выражение:
	// 1. Жанр: все до первого "_"
	// 2. Автор: либо текст между "_" и " — ", либо пусто
	// 3. Название: все после " — "
	const pattern = `^([^_]+)_([^—]*) — (.+)$`
	matches := regexp.MustCompile(pattern).FindStringSubmatch(baseName)

	if len(matches) == 4 {
		originalGenre := strings.TrimSpace(matches[1])
		author := strings.TrimSpace(matches[2])
		title := strings.TrimSpace(matches[3])

		// Применяем маппинг жанров
		genre := originalGenre
		if genresMap != nil {
			if mapped, ok := genresMap[originalGenre]; ok {
				genre = mapped
			} else {
				// Поиск с учетом тримминга пробелов
				trimmedOriginal := strings.TrimSpace(originalGenre)
				for original, mapped := range genresMap {
					if strings.TrimSpace(original) == trimmedOriginal {
						genre = mapped
						break
					}
				}
			}
		}

		tl.Genre = genre
		tl.Author = author
		tl.Title = title
	} else {
		// Если не соответствует шаблону - используем имя файла как название
		tl.Title = baseName
	}

	// Если жанр не указан - используем имя папки
	if tl.Genre == "" {
		tl.Genre = folder
	}

	return tl
}
