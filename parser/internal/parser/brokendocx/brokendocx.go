package brokendocx

import (
	"archive/zip"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Reader представляет собой структуру для чтения .docx по параграфам.
type Reader struct {
	texts    []string
	index    int
	reBase64 *regexp.Regexp
}

// NewReader создает новый Reader для файла .docx.
func NewReader(filepath string, reBase64 *regexp.Regexp) (*Reader, error) {
	texts, err := ParceBrokenXML(filepath, reBase64)
	if err != nil {
		return nil, err
	}
	return &Reader{texts: texts, index: 0, reBase64: reBase64}, nil
}

// Read читает файл .docx по параграфам.
// Если параграфы в файле закончились, возвращает ошибку io.EOF.
func (r *Reader) Read() (string, error) {
	if r.index >= len(r.texts) {
		return "", io.EOF
	}
	text := r.texts[r.index]
	r.index++
	return text, nil
}

// normalizePath заменяет все обратные слэши на прямые.
func normalizePath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// ParceBrokenXML читает файл .docx и возвращает текст из тегов <w:t> построчно.
// Возвращает ошибку, если файл не удалось прочитать или распарсить.
func ParceBrokenXML(filepath string, reBase64 *regexp.Regexp) ([]string, error) {
	// Открываем .docx как ZIP-архив
	zipReader, err := zip.OpenReader(filepath)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии файла .docx: %v", err)
	}
	defer zipReader.Close()

	// Ищем файл word/document.xml
	var documentXML []byte
	for _, file := range zipReader.File {
		// Нормализуем путь
		normalizedPath := normalizePath(file.Name)
		if normalizedPath == "word/document.xml" {
			fileReader, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("ошибка при открытии файла document.xml: %v", err)
			}
			defer fileReader.Close()

			documentXML, err = io.ReadAll(fileReader)
			if err != nil {
				return nil, fmt.Errorf("ошибка при чтении файла document.xml: %v", err)
			}
			break
		}
	}

	if documentXML == nil {
		return nil, fmt.Errorf("файл word/document.xml не найден в архиве")
	}

	// Регулярное выражение для поиска текста внутри тегов <w:t>
	reText := regexp.MustCompile(`<w:t>(.*?)</w:t>`)

	// // Улучшенное регулярное выражение для поиска base64-кодированных данных
	// reBase64 := regexp.MustCompile(`(?:[A-Za-z0-9+/]{40,}={0,2}|iVBORw0KGgo[^"]+)`)

	// Поиск всех совпадений текста
	matches := reText.FindAllSubmatch(documentXML, -1)

	// Срез для хранения текста
	var texts []string

	// Извлечение текста из совпадений и фильтрация артефактов
	for _, match := range matches {
		if len(match) > 1 {
			text := string(match[1])
			// Удаляем base64-кодированные данные если включен режим cut_base64 и переданно регулярное выражение
			if reBase64 != nil && !reBase64.MatchString(text) {
				texts = append(texts, text)
			}
		}
	}

	return texts, nil
}
