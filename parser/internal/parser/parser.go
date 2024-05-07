package parser

import (
	"context"
	"fmt"
	"github.com/terratensor/library/parser/internal/config"
	"github.com/terratensor/library/parser/internal/library/book"
	"github.com/terratensor/library/parser/internal/library/entry"
	"github.com/terratensor/library/parser/internal/parser/docc"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
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

	r, err := docc.NewReader(fp)
	if err != nil {
		str := fmt.Sprintf("%v, %v", filename, err)
		return fmt.Errorf(str)
	}
	defer r.Close()

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
				str := fmt.Sprintf("%v, %v", filename, err)
				return fmt.Errorf(str)
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

		pars = appendParagraph(b, titleList, position, pars)

		b.Reset()

		position++
		batchSizeCount++

		// Записываем пакетам по batchSize параграфов
		if batchSizeCount == p.cfg.BatchSize-1 {
			err = p.storage.Bulk(ctx, pars)
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
		pars = appendParagraph(b, titleList, position, pars)
	}
	b.Reset()

	// Если batchSizeCount меньше batchSize, то записываем оставшиеся параграфы
	if len(pars) > 0 {
		err = p.storage.Bulk(ctx, pars)
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

func appendParagraph(
	b strings.Builder,
	titleList *book.TitleList,
	position int,
	pars entry.PrepareParagraphs,
) entry.PrepareParagraphs {
	parsedParagraph := entry.Entry{
		Genre:    titleList.Genre,
		Author:   titleList.Author,
		BookName: titleList.Title,
		Text:     b.String(),
		Position: position,
		Length:   utf8.RuneCountInString(b.String()),
	}

	pars = append(pars, parsedParagraph)
	return pars
}
