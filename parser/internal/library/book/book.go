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
		Folder: folder, // Сохраняем имя папки
	}

	const pattern = `([^_]+)_([^—]+) — (.+)`
	matches := regexp.MustCompile(pattern).FindStringSubmatch(baseName)

	if len(matches) > 3 {
		originalGenre := matches[1]
		genre := originalGenre

		// Применяем маппинг жанров
		if genresMap != nil {
			if mapped, ok := genresMap[originalGenre]; ok {
				genre = mapped
			} else {
				// Ищем с триммингом пробелов
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
		tl.Author = matches[2]
		tl.Title = matches[3]
	} else {
		tl.Title = baseName
	}

	// Если автор не установлен - используем имя папки
	if tl.Author == "" {
		tl.Author = folder
	}

	return tl
}
