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
	Index string `json:"index"`
	ID    *int64 `json:"id,omitempty"`
	// Doc   entry.Entry `json:"doc"`
	Doc interface{} `json:"doc"` // Changed to interface{} to support different types
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

	tables := []string{"authors", "categories", "titles"}
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
	// log.Printf("a: %v, b: %v, err: %v", a, b, err)
	return err == nil
}

func createTable(ctx context.Context, engine string, tbl string) error {
	const op = "storage.manticore.createTable"

	settings := fmt.Sprintf(`engine='%v' min_infix_len='3' index_exact_words='1' morphology='stem_en, stem_ru, libstemmer_ru, libstemmer_en' index_sp='1' blend_mode='trim_none, skip_pure' blend_chars='-, _, @, &' expand_keywords='1' overshort_step='0' min_stemming_len='4'`, engine)

	var query string
	switch tbl {
	case "authors":
		query = fmt.Sprintf(`create table %v(name string attribute indexed, entry_type string, role string, description text, avatar_file string attribute indexed, created_at timestamp, updated_at timestamp) %v`, tbl, settings)
	case "categories":
		query = fmt.Sprintf(`create table %v(name string attribute indexed, entry_type string, description text, created_at timestamp, updated_at timestamp) %v`, tbl, settings)
	case "titles":
		query = fmt.Sprintf(`create table %v(title string attribute indexed, entry_type string, description text, created_at timestamp, updated_at timestamp) %v`, tbl, settings)
	default:
		query = fmt.Sprintf(`create table %v(source_uuid string, source string attribute indexed, genre string attribute indexed, author string attribute indexed, title string attribute indexed, content text, language string, chunk int, char_count int, word_count int, ocr_quality float, datetime timestamp, created_at timestamp, updated_at timestamp) %v`, tbl, settings)
	}

	sqlRequest := apiClient.UtilsAPI.Sql(ctx).Body(query)
	_, _, err := apiClient.UtilsAPI.SqlExecute(sqlRequest)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// func (c *Client) Bulk(ctx context.Context, entries *[]entry.Entry) error {
// 	const op = "storage.manticore.Bulk"

// 	var body strings.Builder
// 	for _, e := range *entries {
// 		jsonStr, err := json.Marshal(Root{
// 			Insert: Insert{
// 				Index: c.Index,
// 				Doc:   e,
// 			},
// 		})

// 		if err != nil {
// 			return fmt.Errorf("%s: %w", op, err)
// 		}

// 		body.WriteString(string(jsonStr))
// 		body.WriteString(",\n")
// 	}

// 	const maxRetries = 1000
// 	const retryDelay = 100 * time.Millisecond

// 	for attempt := 0; attempt < maxRetries; attempt++ {
// 		_, _, err := c.apiClient.IndexAPI.Bulk(ctx).Body(body.String()).Execute()
// 		if err == nil {
// 			if attempt > 0 {
// 				log.Printf("Successfully inserted data into Manticore after %d attempts", attempt+1)
// 			}
// 			break
// 		}
// 		if attempt < maxRetries-1 {
// 			log.Printf("Failed to insert data into Manticore, retrying... (attempt %d/%d)", attempt+1, maxRetries)
// 			log.Printf("Error: %v", err)
// 			time.Sleep(retryDelay)
// 			continue
// 		}

// 	}

// 	return nil
// }

func (c *Client) Bulk(ctx context.Context, docs interface{}) error {
	const op = "storage.manticore.Bulk"

	var body strings.Builder
	var indexName string

	switch v := docs.(type) {
	case *[]entry.Entry:
		indexName = c.Index
		for _, e := range *v {
			jsonStr, err := json.Marshal(Root{
				Insert: Insert{
					Index: indexName,
					Doc:   e,
				},
			})
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			body.WriteString(string(jsonStr))
			body.WriteString(",\n")
		}
	case *[]entry.Author:
		indexName = "authors"
		for _, a := range *v {
			jsonStr, err := json.Marshal(Root{
				Insert: Insert{
					Index: indexName,
					Doc:   a,
				},
			})
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			body.WriteString(string(jsonStr))
			body.WriteString(",\n")
		}
	case *[]entry.Category:
		indexName = "categories"
		for _, cat := range *v {
			jsonStr, err := json.Marshal(Root{
				Insert: Insert{
					Index: indexName,
					Doc:   cat,
				},
			})
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			body.WriteString(string(jsonStr))
			body.WriteString(",\n")
		}
	case *[]entry.Title:
		indexName = "titles"
		for _, t := range *v {
			jsonStr, err := json.Marshal(Root{
				Insert: Insert{
					Index: indexName,
					Doc:   t,
				},
			})
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			body.WriteString(string(jsonStr))
			body.WriteString(",\n")
		}
	default:
		return fmt.Errorf("%s: unsupported type %T", op, v)
	}

	const maxRetries = 1000
	const retryDelay = 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		_, _, err := c.apiClient.IndexAPI.Bulk(ctx).Body(body.String()).Execute()
		if err == nil {
			if attempt > 0 {
				log.Printf("Successfully inserted data into %s after %d attempts", indexName, attempt+1)
			}
			break
		}
		if attempt < maxRetries-1 {
			log.Printf("Failed to insert data into %s, retrying... (attempt %d/%d)", indexName, attempt+1, maxRetries)
			log.Printf("Error: %v", err)
			time.Sleep(retryDelay)
			continue
		}
		return fmt.Errorf("%s: failed after %d attempts: %w", op, maxRetries, err)
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
