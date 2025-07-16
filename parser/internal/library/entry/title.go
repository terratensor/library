package entry

import "github.com/terratensor/library/parser/internal/library/book"

type Title struct {
	ID          *int64 `json:"id,omitempty"`
	Title       string `json:"title"`
	EntryType   string `json:"entry_type"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

func NewTitle(title string, entryType string) *Title {
	return &Title{
		Title:     title,
		EntryType: entryType,
	}
}

func NewTitleFromTitleList(titleList *book.TitleList) *Title {
	return &Title{
		Title:     titleList.Title,
		EntryType: "book",
	}
}

func (t *Title) SetDescription(description string) {
	t.Description = description
}
