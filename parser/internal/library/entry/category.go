package entry

import "github.com/terratensor/library/parser/internal/library/book"

type Category struct {
	ID          *int64 `json:"id,omitempty"`
	Name        string `json:"name"`
	EntryType   string `json:"entry_type"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

func NewCategory(name string, entryType string) *Category {
	return &Category{
		Name:      name,
		EntryType: entryType,
	}
}

func NewCategoryFromTitleList(titleList *book.TitleList) *Category {
	return &Category{
		Name:      titleList.Genre,
		EntryType: "book",
	}
}

func (c *Category) SetDescription(description string) {
	c.Description = description
}
