package entry

import "context"

type Entry struct {
	ID       *int64
	Genre    string
	Author   string
	BookName string
	Text     string
	Position int
	Length   int
}

type StorageInterface interface {
	// Bulk index operations Post/bulk
	Bulk(ctx context.Context, entries []Entry) error
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
	return nil
}
