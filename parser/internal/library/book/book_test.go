package book

import "testing"

func TestNewTitleList(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		wantGenre  string
		wantAuthor string
		wantTitle  string
	}{
		{
			name:       "no author",
			filePath:   "/books/Альтернативная медицина_ — Прибор частотно-резонансной терапии.docx",
			wantGenre:  "Альтернативная медицина",
			wantAuthor: "",
			wantTitle:  "Прибор частотно-резонансной терапии",
		},
		{
			name:       "with author",
			filePath:   "/books/Фантастика_Иванов — Космическая одиссея.docx",
			wantGenre:  "Фантастика",
			wantAuthor: "Иванов",
			wantTitle:  "Космическая одиссея",
		},
		{
			name:       "with genre",
			filePath:   "/books/Иванов — Космическая одиссея.docx",
			wantGenre:  "books", // будет заменено на имя папки
			wantAuthor: "",
			wantTitle:  "Иванов — Космическая одиссея",
		},
		// Добавьте другие тестовые случаи
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := NewTitleList(tt.filePath, nil, nil)
			if tl.Genre != tt.wantGenre {
				t.Errorf("Genre = %v, want %v", tl.Genre, tt.wantGenre)
			}
			if tl.Author != tt.wantAuthor && tt.wantAuthor != "" {
				t.Errorf("Author = %v, want %v", tl.Author, tt.wantAuthor)
			}
			if tl.Title != tt.wantTitle {
				t.Errorf("Title = %v, want %v", tl.Title, tt.wantTitle)
			}
		})
	}
}
