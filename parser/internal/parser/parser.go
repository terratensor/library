package parser

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/terratensor/library/parser/internal/config"
	"github.com/terratensor/library/parser/internal/library/book"
	"github.com/terratensor/library/parser/internal/library/entry"
	"github.com/terratensor/library/parser/internal/parser/brokendocx"
	"github.com/terratensor/library/parser/internal/parser/docc"
)

type Parser struct {
	cfg      *config.Config
	storage  *entry.Entries
	reBase64 *regexp.Regexp
}

// Глобальная переменная для хранения скомпилированного регулярного выражения
var reBase64 *regexp.Regexp

// Определяем интерфейс, который будет описывать методы, которые используются в docc.Reader и brokendocx.Reader.
type Reader interface {
	Read() (string, error)
}

func NewParser(cfg *config.Config, storage *entry.Entries) *Parser {

	if cfg.Filters.CutBase64 {
		// Улучшенное регулярное выражение для поиска base64-кодированных данных
		reBase64 = regexp.MustCompile(`(?:[A-Za-z0-9+/]{40,}={0,2}|iVBORw0KGgo[^"]+)`)
	}

	return &Parser{
		cfg:      cfg,
		storage:  storage,
		reBase64: reBase64,
	}
}

func (p *Parser) Parse(ctx context.Context, n int, file os.DirEntry, path string) error {

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fp := filepath.Clean(fmt.Sprintf("%v%v", path, file.Name()))

	var filename = file.Name()
	var extension = filepath.Ext(filename)
	var bookName = filename[0 : len(filename)-len(extension)]

	titleList := book.NewTitleList(bookName)
	titleList.SourceUUID = uuid.New()
	titleList.Source = filename

	r, err := docc.NewReader(fp, p.reBase64)
	if err != nil {
		return fmt.Errorf("%v, %v", filename, err)
	}
	defer r.Close()

	// Парсим текст из docx файла, если получим ошибку, то надо запустить brokenXML парсер
	err = p.runBuilder(ctx, r, filename, titleList)
	if err != nil {
		// Если установлен режим в конфигурации BrokenDocxMode, то запускаем дополнительный билдер
		if p.cfg.BrokenDocxMode {
			log.Println(err)
			br, err := brokendocx.NewReader(fp, p.reBase64)
			if err != nil {
				return fmt.Errorf("ошибка при создании Reader: %v", err)
			}
			// запускаем дополнительный билдер
			log.Printf("Запускаем дополнительный билдер: %v", filename)
			err = p.runBuilder(ctx, br, filename, titleList)
			if err != nil {
				return fmt.Errorf("%v, %v", filename, err)
			}
		} else {
			log.Println("Для обработки сломанных документов включите режим BrokenDocxMode в конфигурации.")
			return fmt.Errorf("%v, %v", filename, err)
		}
	}
	//log.Printf("%v #%v done", newBook.Filename, n+1)
	return nil
}

func (p *Parser) splitLongParagraph(longBuilder *strings.Builder, builder *strings.Builder) {
	result := longBuilder.String()
	result = strings.TrimPrefix(result, "<div>")
	result = strings.TrimSuffix(result, "</div>")

	// sentences []string Делим параграф на предложения, разделитель точка с пробелом
	sentences := strings.SplitAfter(result, ".")
	//sentences := regexp.MustCompile(`[.!?]`).Split(result, -1)

	longBuilder.Reset()

	var flag bool

	for n, sentence := range sentences {

		sentence = strings.TrimSpace(sentence)
		if n == 0 {
			builder.WriteString("<div>")
		}
		if (utf8.RuneCountInString(builder.String()) + utf8.RuneCountInString(sentence)) < p.cfg.OptParSize {

			builder.WriteString(sentence)
			builder.WriteString(" ")
			continue
		}
		if !flag {
			builder.WriteString(strings.TrimSpace(sentence))
			builder.WriteString("</div>")
			flag = true
			if len(sentences) == n+1 {
				break
			}
			longBuilder.WriteString("<div>")

			continue
		}

		longBuilder.WriteString(sentence)
		longBuilder.WriteString(" ")

	}
	if utf8.RuneCountInString(longBuilder.String()) > 0 {
		temp := longBuilder.String()
		longBuilder.Reset()
		longBuilder.WriteString(strings.TrimSpace(temp))
		longBuilder.WriteString("</div>")
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
		Text:       text,
		Position:   position,
		Length:     utf8.RuneCountInString(b.String()),
	}

	pars = append(pars, parsedParagraph)
	return pars
}

func (p *Parser) runBuilder(ctx context.Context, r Reader, filename string, titleList *book.TitleList) error {

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
