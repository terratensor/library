package book

import (
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
}

// NewTitleList creates a new TitleList from a string
func NewTitleList(str string, genresMap map[string]string) *TitleList {
	const pattern = `([^_]+)_([^—]+) — (.+)`
	matches := regexp.MustCompile(pattern).FindStringSubmatch(str)
	if len(matches) > 3 {
		originalGenre := matches[1]
		genre := originalGenre // По умолчанию оставляем оригинальный жанр

		// Применяем маппинг жанров
		if genresMap != nil {
			// Ищем полное совпадение
			if mapped, ok := genresMap[originalGenre]; ok {
				genre = mapped
			} else {
				// Если полного совпадения нет, ищем с учетом тримминга пробелов
				for original, mapped := range genresMap {
					if strings.TrimSpace(originalGenre) == strings.TrimSpace(original) {
						genre = mapped
						break
					}
				}
			}
		}

		return &TitleList{
			Genre:  genre,
			Author: matches[2],
			Title:  matches[3],
		}
	}
	return &TitleList{Title: str}
}
