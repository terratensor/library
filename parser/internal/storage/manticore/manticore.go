package manticore

import (
	"context"
	"encoding/json"
	"fmt"
	openapiclient "github.com/manticoresoftware/manticoresearch-go"
	"github.com/terratensor/library/parser/internal/config"
	"github.com/terratensor/library/parser/internal/library/entry"
	"strings"
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

	tbl := cfg.Index
	engine := cfg.Engine
	// Check if table exists in cache
	//exists := tableExists(ctx, tbl)
	//if !exists {
	// Create table if it doesn't exist
	if err := createTable(ctx, engine, tbl); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	//}

	return &Client{Index: tbl, apiClient: apiClient}, nil
}

func tableExists(ctx context.Context, tbl string) bool {
	_, _, err := apiClient.UtilsAPI.Sql(ctx).Body(fmt.Sprintf("SHOW CREATE TABLE %v", tbl)).Execute()
	if err != nil {
		return false
	}
	return true
}

func createTable(ctx context.Context, engine string, tbl string) error {
	const op = "storage.manticore.createTable"

	query := fmt.Sprintf("create table %v(genre text, author text, title text, `text` text, position int, length int) engine='%v' min_infix_len='3' index_exact_words='1' morphology='stem_en, stem_ru' index_sp='1'", tbl, engine)

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

	_, _, err := c.apiClient.IndexAPI.Bulk(ctx).Body(body.String()).Execute()

	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
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
