package parser

import (
	"archive/tar"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/terratensor/library/parser/internal/config"
	"github.com/terratensor/library/parser/internal/library/book"
	"github.com/terratensor/library/parser/internal/library/entry"
	"github.com/terratensor/library/parser/internal/metadata"
	"github.com/terratensor/library/parser/internal/parser/brokendocx"
	"github.com/terratensor/library/parser/internal/parser/docc"
)

type Parser struct {
	cfg      *config.Config
	storage  *entry.Entries
	reBase64 *regexp.Regexp
	// Add these fields to track unique models
	authors    map[string]entry.Author
	categories map[string]entry.Category
	titles     map[string]entry.Title
	mu         sync.Mutex        // To protect concurrent access to maps
	genresMap  map[string]string // Маппинг жанров
}

// Глобальная переменная для хранения скомпилированного регулярного выражения
var reBase64 *regexp.Regexp

// Определяем интерфейс, который будет описывать методы, которые используются в docc.Reader и brokendocx.Reader.
type Reader interface {
	Read() (string, error)
}

// FileInfo содержит информацию о файле для обработки
type FileInfo struct {
	TempPath  string // временный путь к файлу
	OrigName  string // оригинальное имя файла
	Extension string // расширение файла
}

func NewParser(cfg *config.Config, storage *entry.Entries) *Parser {
	if cfg.Filters.CutBase64 {
		// Улучшенное регулярное выражение для поиска base64-кодированных данных
		reBase64 = regexp.MustCompile(`(?:[A-Za-z0-9+/]{40,}={0,2}|iVBORw0KGgo[^"]+)`)
	}

	// Загружаем маппинг жанров из CSV
	genresMap := make(map[string]string)
	if cfg.GenresMapPath != "" {
		file, err := os.Open(cfg.GenresMapPath)
		if err != nil {
			log.Printf("Warning: could not open genres map file: %v", err)
		} else {
			defer file.Close()

			reader := csv.NewReader(file)
			reader.Comma = ','
			reader.FieldsPerRecord = 2 // Ожидаем ровно 2 колонки
			reader.LazyQuotes = true   // Для обработки строк в кавычках

			records, err := reader.ReadAll()
			if err != nil {
				log.Printf("Warning: could not read genres map file: %v", err)
			} else {
				for _, record := range records {
					if len(record) == 2 {
						original := strings.TrimSpace(record[0])
						mapped := strings.TrimSpace(record[1])
						genresMap[original] = mapped
					}
				}
			}
		}
	}

	return &Parser{
		cfg:        cfg,
		storage:    storage,
		reBase64:   reBase64,
		authors:    make(map[string]entry.Author),
		categories: make(map[string]entry.Category),
		titles:     make(map[string]entry.Title),
		genresMap:  genresMap,
	}
}

// ProcessTar обрабатывает tar-архив параллельно
func (p *Parser) ProcessTar(ctx context.Context, tarStream io.Reader, workers int) error {
	tr := tar.NewReader(tarStream)
	tasks := make(chan FileInfo, workers)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	// Запускаем worker-ов
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				select {
				case <-ctx.Done():
					return
				default:
					if err := p.processFile(ctx, task); err != nil {
						select {
						case errCh <- err:
						default:
						}
					}
				}
			}
		}()
	}

	defer close(tasks)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		default:
			hdr, err := tr.Next()
			if err == io.EOF {
				wg.Wait()
				return nil
			}
			if err != nil {
				return err
			}

			if hdr.Typeflag == tar.TypeReg {
				ext := filepath.Ext(hdr.Name)
				if ext != ".docx" && ext != ".pdf" && ext != ".epub" {
					continue
				}

				tmpFile, err := os.CreateTemp("", "doc_*.tmp")
				if err != nil {
					return err
				}

				if _, err := io.Copy(tmpFile, tr); err != nil {
					tmpFile.Close()
					os.Remove(tmpFile.Name())
					return err
				}
				tmpFile.Close()

				tasks <- FileInfo{
					TempPath:  tmpFile.Name(),
					OrigName:  hdr.Name,
					Extension: ext,
				}
			}
		}
	}
}

// processFile обрабатывает один файл из архива
func (p *Parser) processFile(ctx context.Context, info FileInfo) error {
	defer os.Remove(info.TempPath)

	file := &dirEntry{
		name:  filepath.Base(info.TempPath),
		isDir: false,
	}

	// Модифицированный Parse с поддержкой оригинального имени
	return p.ParseWithOrigName(ctx, file, filepath.Dir(info.TempPath), info.OrigName)
}

// ParseWithOrigName - модифицированная версия Parse с поддержкой оригинального имени
func (p *Parser) ParseWithOrigName(ctx context.Context, file os.DirEntry, path, origName string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fp := filepath.Clean(filepath.Join(path, file.Name()))
	filename := origName // Используем оригинальное имя
	extension := strings.ToLower(filepath.Ext(filename))
	bookName := filename[:len(filename)-len(extension)]

	titleList := book.NewTitleList(bookName, p.genresMap)
	titleList.SourceUUID = uuid.New()
	titleList.Source = filename

	switch extension {
	case ".docx":
		return p.parseDocx(ctx, fp, filename)
	case ".pdf":
		return p.parsePDF(ctx, fp, filename)
	case ".epub":
		return p.parseEPUB(ctx, fp, filename)
	default:
		return fmt.Errorf("unsupported file format: %s", extension)
	}
}

// dirEntry реализует os.DirEntry для временных файлов
type dirEntry struct {
	name  string
	isDir bool
}

func (d *dirEntry) Name() string               { return d.name }
func (d *dirEntry) IsDir() bool                { return d.isDir }
func (d *dirEntry) Type() os.FileMode          { return 0 }
func (d *dirEntry) Info() (os.FileInfo, error) { return nil, nil }

func (p *Parser) processModels(ctx context.Context, titleList *book.TitleList) error {
	// Process author
	if titleList.Author != "" {
		p.mu.Lock()
		if _, exists := p.authors[titleList.Author]; !exists {
			author := entry.NewAuthorFromTitleList(titleList)
			p.authors[titleList.Author] = *author
		}
		p.mu.Unlock()
	}

	// Process category
	if titleList.Genre != "" {
		p.mu.Lock()
		if _, exists := p.categories[titleList.Genre]; !exists {
			category := entry.NewCategoryFromTitleList(titleList)
			p.categories[titleList.Genre] = *category
		}
		p.mu.Unlock()
	}

	// Process title
	if titleList.Title != "" {
		p.mu.Lock()
		if _, exists := p.titles[titleList.Title]; !exists {
			title := entry.NewTitleFromTitleList(titleList)
			p.titles[titleList.Title] = *title
		}
		p.mu.Unlock()
	}

	return nil
}

// Add this new method to store all collected models
func (p *Parser) StoreModels(ctx context.Context, mp *metadata.Processor) error {
	authors := mp.GetAuthors()
	categories := mp.GetCategories()
	titles := mp.GetTitles()

	// Сохраняем в базу
	if len(authors) > 0 {
		if err := p.storage.BulkAuthors(ctx, authors); err != nil {
			return fmt.Errorf("failed to store authors: %v", err)
		}
	}

	if len(categories) > 0 {
		if err := p.storage.BulkCategories(ctx, categories); err != nil {
			return fmt.Errorf("failed to store categories: %v", err)
		}
	}

	if len(titles) > 0 {
		if err := p.storage.BulkTitles(ctx, titles); err != nil {
			return fmt.Errorf("failed to store titles: %v", err)
		}
	}

	return nil
}

func (p *Parser) Parse(ctx context.Context, file os.DirEntry, path string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fp := filepath.Clean(filepath.Join(path, file.Name()))
	filename := file.Name()
	extension := strings.ToLower(filepath.Ext(filename))
	bookName := filename[:len(filename)-len(extension)]

	titleList := book.NewTitleList(bookName, p.genresMap)
	titleList.SourceUUID = uuid.New()
	titleList.Source = filename

	// Обрабатываем модели с блокировкой
	// p.mu.Lock()
	// if titleList.Author != "" {
	// 	if _, exists := p.authors[titleList.Author]; !exists {
	// 		author := entry.NewAuthorFromTitleList(titleList)
	// 		p.authors[titleList.Author] = *author
	// 	}
	// }
	// if titleList.Genre != "" {
	// 	if _, exists := p.categories[titleList.Genre]; !exists {
	// 		category := entry.NewCategoryFromTitleList(titleList)
	// 		p.categories[titleList.Genre] = *category
	// 	}
	// }
	// if titleList.Title != "" {
	// 	if _, exists := p.titles[titleList.Title]; !exists {
	// 		title := entry.NewTitleFromTitleList(titleList)
	// 		p.titles[titleList.Title] = *title
	// 	}
	// }
	// p.mu.Unlock()

	switch extension {
	case ".docx":
		return p.parseDocx(ctx, fp, filename)
	case ".pdf":
		return p.parsePDF(ctx, fp, filename)
	case ".epub":
		return p.parseEPUB(ctx, fp, filename)
	default:
		return fmt.Errorf("unsupported file format: %s", extension)
	}
}

func (p *Parser) parseDocx(ctx context.Context, filePath, filename string) error {
	titleList := book.NewTitleList(strings.TrimSuffix(filename, filepath.Ext(filename)), p.genresMap)
	titleList.SourceUUID = uuid.New()
	titleList.Source = filename

	// Пытаемся обработать как обычный docx
	r, err := docc.NewReader(filePath, p.reBase64)
	if err != nil {
		return fmt.Errorf("%v, %v", filename, err)
	}
	defer r.Close()

	err = p.runBuilder(ctx, r, filename, titleList)
	if err != nil {
		if p.cfg.BrokenDocxMode {
			log.Printf("Failed to parse as normal DOCX, trying broken DOCX parser: %v", err)
			return p.parseBrokenDocx(ctx, filePath, filename, titleList)
		}
		return fmt.Errorf("%v, %v", filename, err)
	}
	return nil
}

func (p *Parser) parseBrokenDocx(ctx context.Context, filePath, filename string, titleList *book.TitleList) error {
	br, err := brokendocx.NewReader(filePath, p.reBase64)
	if err != nil {
		return fmt.Errorf("failed to create broken DOCX reader: %v", err)
	}
	defer br.Close()

	log.Printf("Using broken DOCX parser for: %v", filename)
	err = p.runBuilder(ctx, br, filename, titleList)
	if err != nil {
		return fmt.Errorf("broken DOCX parser failed for %v: %v", filename, err)
	}
	return nil
}

func (p *Parser) parsePDF(ctx context.Context, filePath, filename string) error {
	if !p.cfg.PDFMode {
		return fmt.Errorf("PDF processing is disabled in config")
	}

	titleList := book.NewTitleList(strings.TrimSuffix(filename, filepath.Ext(filename)), p.genresMap)
	titleList.SourceUUID = uuid.New()
	titleList.Source = filename

	// Заглушка - возвращаем ошибку, что функционал еще не реализован
	return fmt.Errorf("PDF parser is not implemented yet")
}

func (p *Parser) parseEPUB(ctx context.Context, filePath, filename string) error {
	if !p.cfg.EPUBMode {
		return fmt.Errorf("EPUB processing is disabled in config")
	}

	titleList := book.NewTitleList(strings.TrimSuffix(filename, filepath.Ext(filename)), p.genresMap)
	titleList.SourceUUID = uuid.New()
	titleList.Source = filename

	// Заглушка - возвращаем ошибку, что функционал еще не реализован
	return fmt.Errorf("EPUB parser is not implemented yet")
}

func (p *Parser) runBuilder(ctx context.Context, r Reader, filename string, titleList *book.TitleList) error {

	// Process models first
	if err := p.processModels(ctx, titleList); err != nil {
		return err
	}

	// position номер параграфа в индексе
	position := 1

	var pars entry.PrepareParagraphs

	// var b билдер
	// var textBuilder билдер для текста прочитанного из docx файла
	// var bufBuilder промежуточный билдер для текста, для соединения параграфов
	// var longParBuilder билдер в котором текущий обрабатываемый длинный параграф
	var b,
		textBuilder,
		bufBuilder,
		longParBuilder strings.Builder

	batchSizeCount := 0
	for {
		// Используем select для выхода по истечении контекста, прерывание выполнения ctrl+c
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// если билдер длинного параграфа пуст и буфер текста пуст,
		// то читаем следующий параграф из файла docx и передаем его в textBuilder
		if utf8.RuneCountInString(longParBuilder.String()) == 0 && utf8.RuneCountInString(textBuilder.String()) == 0 {
			text, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("%v, %v", filename, err)
			}
			// Если строка пустая, то пропускаем
			// и переходим к следующей итерации цикла
			if text == "" {
				continue
			}
			// обрабатываем троеточия в получившемся оптимальном параграфе
			text = processTriples(text)
			// Если кол-во символов в тексте больше максимально установленной длины,
			// записываем текст в буфер большого параграфа, иначе записываем текст в текстовый буфер
			if utf8.RuneCountInString(text) > p.cfg.MaxParSize {
				longParBuilder.WriteString(text)
			} else {
				textBuilder.WriteString(text)
			}
		}

		// запись остатка от длинного параграфа в обычный билдер при условии, что остаток менее maxParSize
		if utf8.RuneCountInString(longParBuilder.String()) > 0 && utf8.RuneCountInString(longParBuilder.String()) < p.cfg.MaxParSize {
			b.WriteString(longParBuilder.String())
			longParBuilder.Reset()
		}
		// Если кол-во символов текста в билдер буфере большого параграфа больше максимальной границы maxParSize
		// разбиваем параграф на 2 части, оптимальной длины и остаток,
		// остаток сохраняем в longParBuilder, оптимальную часть сохраняем в builder b
		if utf8.RuneCountInString(longParBuilder.String()) >= p.cfg.MaxParSize {
			p.splitLongParagraph(&longParBuilder, &b)
		}

		// Если в билдер-буфере есть записанный параграф, то записываем его в обычный билдер b и очищаем билдер-буфер
		if utf8.RuneCountInString(bufBuilder.String()) > 0 {
			if utf8.RuneCountInString(bufBuilder.String()) >= p.cfg.MaxParSize {
				log.Println("stage 6")
				log.Printf("в билдер буфере длинный параграф %v\r\n", utf8.RuneCountInString(bufBuilder.String()))
				panic("panic")
			}
			b.WriteString(bufBuilder.String())
			bufBuilder.Reset()
		}

		// Кол-во символов в билдере, получено от предыдущей или текущей итерации
		builderLength := utf8.RuneCountInString(b.String())

		// Кол-во символов в текущем обрабатываемом параграфе, получено из парсера
		textLength := utf8.RuneCountInString(textBuilder.String())

		// Сумма кол-ва символов в предыдущих склеенных и в текущем параграфах
		concatLength := builderLength + textLength

		// Если кол-во символов в результирующей строке concatLength менее
		// минимального значения длины параграфа minParSize,
		// то соединяем предыдущие параграфы и текущий обрабатываемый,
		// переходим к следующей итерации цикла и читаем следующий параграф из файла docx,
		// повторяем пока не достигнем границы минимального значения длины параграфа

		// и нет длинного параграфа в обработке
		if concatLength < p.cfg.MinParSize && utf8.RuneCountInString(longParBuilder.String()) == 0 {
			b.WriteString(textBuilder.String())
			textBuilder.Reset()
			continue
		}
		// Если кол-во символов в результирующей строке билдера более или равно
		// минимальному значению длины параграфа mixParSize и менее или равно
		// оптимальному значению длины параграфа, то переходим к следующей итерации цикла
		// и читаем следующий параграф из файла docx

		// и нет длинного параграфа в обработке
		if concatLength >= p.cfg.MinParSize &&
			float64(concatLength) <= float64(p.cfg.OptParSize)*1.05 &&
			utf8.RuneCountInString(longParBuilder.String()) == 0 {

			b.WriteString(textBuilder.String())
			textBuilder.Reset()
			continue
		}

		if concatLength > p.cfg.OptParSize && concatLength <= p.cfg.MaxParSize {
			b.WriteString(textBuilder.String())
			textBuilder.Reset()
		}

		pars = appendParagraph(b, titleList, position, pars, p.cfg.Filters.CutBase64Recursive)

		b.Reset()

		position++
		batchSizeCount++

		// Записываем пакетам по batchSize параграфов
		if batchSizeCount == p.cfg.BatchSize-1 {
			err := p.storage.Bulk(ctx, pars)
			if err != nil {
				log.Printf("log bulk insert error query: %v \r\n", err)
			}
			// очищаем slice
			pars = nil
			batchSizeCount = 0
		}

	}

	// Если билдер строки не пустой, записываем оставшийся текст в параграфы и сбрасываем билдер
	if utf8.RuneCountInString(b.String()) > 0 {
		pars = appendParagraph(b, titleList, position, pars, p.cfg.Filters.CutBase64Recursive)
	}
	b.Reset()

	// Если batchSizeCount меньше batchSize, то записываем оставшиеся параграфы
	if len(pars) > 0 {
		err := p.storage.Bulk(ctx, pars)
		if err != nil {
			log.Printf("log bulk insert error query: %v \r\n", err)
		}
	}

	return nil
}

func (p *Parser) splitLongParagraph(longBuilder *strings.Builder, builder *strings.Builder) {
	result := longBuilder.String()
	// result = strings.TrimPrefix(result, "<div>")
	// result = strings.TrimSuffix(result, "</div>")

	// sentences []string Делим параграф на предложения, разделитель точка с пробелом
	// sentences := strings.SplitAfter(result, ".")
	//sentences := regexp.MustCompile(`[.!?]`).Split(result, -1)

	// Используем улучшенную функцию, для разбиения параграфа на предложения
	sentences := splitMessageOnSentences(result)

	longBuilder.Reset()

	var flag bool

	for n, sentence := range sentences {

		sentence = strings.TrimSpace(sentence)
		// if n == 0 {
		// 	builder.WriteString("<div>")
		// }
		if (utf8.RuneCountInString(builder.String()) + utf8.RuneCountInString(sentence)) < p.cfg.OptParSize {

			builder.WriteString(sentence)
			builder.WriteString(" ")
			continue
		}
		if !flag {
			builder.WriteString(strings.TrimSpace(sentence))
			// builder.WriteString("</div>")
			builder.WriteString("\n\n")
			flag = true
			if len(sentences) == n+1 {
				break
			}
			// longBuilder.WriteString("<div>")

			continue
		}

		longBuilder.WriteString(sentence)
		longBuilder.WriteString(" ")

	}

	if utf8.RuneCountInString(longBuilder.String()) > 0 {
		temp := longBuilder.String()
		longBuilder.Reset()
		longBuilder.WriteString(strings.TrimSpace(temp))
		// longBuilder.WriteString("</div>")
		longBuilder.WriteString("\n\n")
	}
}

// processTriples функция обработки троеточий в итоговом спарсенном параграфе,
// приводит все троеточия к виду …
func processTriples(text string) string {
	text = strings.Replace(text, ". . .", "…", -1)
	text = strings.Replace(text, "...", "…", -1)
	return text
}

func appendParagraph(b strings.Builder, titleList *book.TitleList, position int, pars entry.PrepareParagraphs, cutBase64Recursive bool) entry.PrepareParagraphs {

	text := b.String()
	// Если установлен режмим в конфигурации RecursiveCutBase64, то вырезаем все base64 данные из получившегося параграфа
	if cutBase64Recursive {
		// Запускаем функцию, которая рекурсивно вырезает все base64 данные из получившегося параграфа
		text = recursiveCutBase64(text)
	}
	parsedParagraph := entry.Entry{
		SourceUUID: titleList.SourceUUID,
		Source:     titleList.Source,
		Genre:      titleList.Genre,
		Author:     titleList.Author,
		BookName:   titleList.Title,
		Content:    text,
		Chunk:      position,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}

	parsedParagraph.CalculateCharCount()
	parsedParagraph.CalculateWordCount()
	parsedParagraph.DetectLanguage()
	parsedParagraph.CalculateOCRQuality()

	// log.Printf("parsedParagraph: %v", parsedParagraph)
	// panic("stop")

	pars = append(pars, parsedParagraph)
	return pars
}

// Функция для рекурсивного вырезания совпадений по регулярному выражению
func recursiveCutBase64(input string) string {
	// Компилируем регулярное выражение
	reBase64 := regexp.MustCompile(`(?:[A-Za-z0-9+/]{40,}={0,2}|iVBORw0KGgo[^"]+)`)

	// Если совпадений нет, возвращаем исходную строку
	if !reBase64.MatchString(input) {
		return input
	}

	// Вырезаем все совпадения из строки
	input = reBase64.ReplaceAllString(input, "")

	// Рекурсивно вызываем функцию для оставшейся строки
	return recursiveCutBase64(input)
}

// splitMessageOnSentences разделяет текст на предложения по знакам препинания.
// Функция принимает строку `chunk` и возвращает слайс строк, где каждый элемент — это отдельное предложение.
func splitMessageOnSentences(chunk string) []string {
	// punct — это множество (map) знаков препинания, которые обозначают конец предложения.
	punct := map[rune]struct{}{'.': {}, '!': {}, '?': {}, '…': {}}

	// Разбиваем входной текст на слова с помощью strings.Fields.
	// strings.Fields разделяет строку по пробелам и возвращает слайс слов.
	words := strings.Fields(chunk)

	// result — это слайс, в который будут добавляться готовые предложения.
	var result []string
	// builder используется для построения предложений.
	var builder strings.Builder

	// Проходим по каждому слову в слайсе words.
	for _, word := range words {
		// Определяем последний символ в слове.
		lastRune, _ := utf8.DecodeLastRuneInString(word)

		// Добавляем текущее слово и пробел в builder.
		builder.WriteString(word)
		builder.WriteString(" ")

		// Если последний символ слова является знаком препинания из множества punct,
		// это означает, что предложение закончилось.
		if _, exists := punct[lastRune]; exists {
			// Добавляем собранное предложение в result, удаляя лишние пробелы с помощью strings.TrimSpace.
			result = append(result, strings.TrimSpace(builder.String()))
			// Сбрасываем builder для построения следующего предложения.
			builder.Reset()
		}
	}

	// Если после завершения цикла в builder остались данные,
	// это означает, что последнее предложение не завершено знаком препинания.
	// Добавляем его в result.
	if builder.Len() > 0 {
		result = append(result, strings.TrimSpace(builder.String()))
	}
	// if len(result) > 0 {
	// 	return splitBlocks(result, msgsign, " ", limit)
	// }

	// Возвращаем слайс предложений.
	return result
}

// ProcessMetadataOnly обрабатывает только метаданные файлов
func (p *Parser) ProcessMetadataOnly(ctx context.Context, mp *metadata.Processor, file os.DirEntry, path string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fullPath := filepath.Join(path, file.Name())
	return mp.ProcessFile(fullPath)
}
