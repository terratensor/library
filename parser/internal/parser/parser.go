package parser

import (
	"context"
	"github.com/terratensor/library/parser/internal/config"
	"github.com/terratensor/library/parser/internal/library/entry"
	"os"
)

type Parser struct {
	cfg     *config.Config
	storage *entry.Entries
}

func NewParser(cfg *config.Config, storage *entry.Entries) *Parser {
	return &Parser{
		cfg:     cfg,
		storage: storage,
	}
}

func (p *Parser) Parse(ctx context.Context, n int, file os.DirEntry, path string) error {
	return nil
}
