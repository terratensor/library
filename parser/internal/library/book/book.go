package book

import (
	"regexp"

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
		genre := matches[1]

		// Применяем маппинг жанров, если он доступен
		if genresMap != nil {
			if mapped, ok := genresMap[genre]; ok {
				genre = mapped
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
