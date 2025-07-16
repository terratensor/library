package entry

import "github.com/terratensor/library/parser/internal/library/book"

type Author struct {
	ID          *int64 `json:"id,omitempty"`
	Name        string `json:"name"`
	EntryType   string `json:"entry_type"`
	Role        string `json:"role"`
	Description string `json:"description"`
	Avatar_file string `json:"avatar_file"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

func NewAuthor(name string, entryType string) *Author {
	return &Author{
		Name:      name,
		EntryType: entryType,
	}
}

func NewAuthorFromTitleList(titleList *book.TitleList) *Author {
	return &Author{
		Name:      titleList.Author,
		EntryType: "book",
	}
}

func (a *Author) SetDescription(description string) {
	a.Description = description
}

func (a *Author) SetAvatarFile(avatarFile string) {
	a.Avatar_file = avatarFile
}

func (a *Author) SetRole(role string) {
	a.Role = role
}
