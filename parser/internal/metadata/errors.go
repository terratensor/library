package metadata

import (
	"fmt"
	"os"

	"github.com/terratensor/library/parser/internal/library/book"
)

type Report struct {
	Duplicates map[string][]string
	Entries    map[string]book.TitleList
	ErrorLog   *os.File
}

func (mp *Processor) GenerateReport() *Report {
	return &Report{
		Duplicates: mp.duplicates,
		Entries:    mp.entries,
		ErrorLog:   mp.errorLog,
	}
}

func (mp *Processor) SaveDuplicatesReport(path string) error {
	if len(mp.duplicates) == 0 {
		mp.logger.Info("no duplicate titles found")
		return nil
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create duplicates report: %v", err)
	}
	defer f.Close()

	f.WriteString("Duplicate Titles Report\n")
	f.WriteString("======================\n\n")

	for title, files := range mp.duplicates {
		if len(files) < 2 {
			continue
		}
		f.WriteString(fmt.Sprintf("Title: %s\n", title))
		f.WriteString(fmt.Sprintf("Found in %d files:\n", len(files)))
		for _, file := range files {
			f.WriteString(fmt.Sprintf(" - %s\n", file))
		}
		f.WriteString("\n")
	}

	mp.logger.Info("duplicates report saved", "path", path)
	return nil
}
