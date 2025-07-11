package manticore

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	openapiclient "github.com/manticoresoftware/manticoresearch-go"
	"github.com/terratensor/library/parser/internal/config"
	"github.com/terratensor/library/parser/internal/library/entry"
)

var _ entry.StorageInterface = &Client{}

var apiClient *openapiclient.APIClient

type Client struct {
	Index     string
	apiClient *openapiclient.APIClient
}

type Insert struct {
	Index string      `json:"index"`
	ID    *int64      `json:"id,omitempty"`
	Doc   entry.Entry `json:"doc"`
}

type Root struct {
	Insert Insert `json:"insert"`
}

func New(ctx context.Context, cfg *config.Manticore) (*Client, error) {
	const op = "storage.manticore.New"
	// Initialize apiClient
	configuration := openapiclient.NewConfiguration()
	configuration.Servers = openapiclient.ServerConfigurations{{URL: serverConfigurationURL(cfg)}}
	apiClient = openapiclient.NewAPIClient(configuration)

	tables := []string{}
	mtbl := cfg.Index
	tables = append(tables, mtbl)
	engine := cfg.Engine

	// Check if table exists in cache
	for _, tbl := range tables {
		exists := tableExists(ctx, tbl)
		if !exists {
			// Create table if it doesn't exist
			if err := createTable(ctx, engine, tbl); err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
		}
	}

	return &Client{Index: mtbl, apiClient: apiClient}, nil
}

// tableExists checks whether a table with the specified name exists in the database.
//
// Parameters:
// - ctx: The context for managing request deadlines and cancellations.
// - tableName: The name of the table to check for existence.
//
// Returns:
// - bool: True if the table exists, false otherwise.
func tableExists(ctx context.Context, tableName string) bool {
	showCreateTableQuery := fmt.Sprintf("SHOW CREATE TABLE %v", tableName)
	showCreateTableRequest := apiClient.UtilsAPI.Sql(ctx).Body(showCreateTableQuery)

	_, _, err := showCreateTableRequest.Execute()
	return err == nil
}

// createTable создает таблицу в Manticore Search со схемой, зависящей от имени таблицы.
// Функция поддерживает специальные схемы для таблиц с именами "genre", "author" или "source",
// и использует схему по умолчанию для всех остальных имен таблиц.
//
// Параметры:
//   - ctx: context.Context - Контекст для отмены операций и таймаутов
//   - engine: string - Движок хранения Manticore для создаваемой таблицы
//   - tbl: string - Имя создаваемой таблицы. Определяет схему:
//   - "genre": создает таблицу с полями genre_id, genre_uuid, name (индексированное) и временными метками
//   - "author": создает таблицу с полями author_id, author_uuid, name (индексированное) и временными метками
//   - "source": создает таблицу с полями source_uuid, name (индексированное) и временными метками
//   - любое другое имя: создает таблицу с общей схемой, включающей поля для
//     genre_id, author_id, title, text, position, length, source_uuid и временные метки
//
// Возвращает:
//   - error: Возвращает nil при успешном создании, или ошибку если создание не удалось.
//     Ошибка будет обернута с именем операции "storage.manticore.createTable".
//
// Особенности:
//   - Все таблицы создаются с оптимизированными для поиска настройками:
//   - min_infix_len='3' - включает поиск по подстрокам для терминов от 3 символов
//   - index_exact_words='1' - индексирует точные формы слов
//   - morphology='stem_en, stem_ru' - включает стемминг для английского и русского
//   - index_sp='1' - индексирует границы предложений и абзацев
func createTable(ctx context.Context, engine string, tbl string) error {
	const op = "storage.manticore.createTable"

	query := fmt.Sprintf("create table %v(source_uuid string, source string indexed attribute, genre string indexed attribute, author string indexed attribute, title string indexed attribute, `text` text, position int, length int, datetime timestamp, created_at timestamp, updated_at timestamp) engine='%v' min_infix_len='3' index_exact_words='1' morphology='stem_en, stem_ru' index_sp='1'", tbl, engine)

	sqlRequest := apiClient.UtilsAPI.Sql(ctx).Body(query)
	_, _, err := apiClient.UtilsAPI.SqlExecute(sqlRequest)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (c *Client) Bulk(ctx context.Context, entries *[]entry.Entry) error {
	const op = "storage.manticore.Bulk"

	var body strings.Builder
	for _, e := range *entries {
		jsonStr, err := json.Marshal(Root{
			Insert: Insert{
				Index: c.Index,
				Doc:   e,
			},
		})

		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		body.WriteString(string(jsonStr))
		body.WriteString(",\n")
	}

	const maxRetries = 1000
	const retryDelay = 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		_, _, err := c.apiClient.IndexAPI.Bulk(ctx).Body(body.String()).Execute()
		if err == nil {
			if attempt > 0 {
				log.Printf("Successfully inserted data into Manticore after %d attempts", attempt+1)
			}
			break
		}
		if attempt < maxRetries-1 {
			log.Printf("Failed to insert data into Manticore, retrying... (attempt %d/%d)", attempt+1, maxRetries)
			log.Printf("Error: %v", err)
			time.Sleep(retryDelay)
			continue
		}

	}

	return nil
}

// serverConfigurationURL generates the server configuration URL based on the provided Manticore configuration.
//
// Parameters:
// - cfg: A pointer to the Manticore configuration struct.
//
// Returns:
// - string: The generated server configuration URL.
func serverConfigurationURL(cfg *config.Manticore) string {
	var builder strings.Builder
	builder.WriteString("http://")
	builder.WriteString(cfg.Host)
	builder.WriteString(":")
	builder.WriteString(cfg.Port)
	return builder.String()
}
