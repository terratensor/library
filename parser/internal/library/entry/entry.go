package entry

import (
	"context"
	"fmt"

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
	Text       string    `json:"text"`
	Position   int       `json:"position"`
	Length     int       `json:"length"`
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
