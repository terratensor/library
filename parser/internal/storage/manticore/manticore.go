package manticore

import (
	"context"
	"fmt"
	openapiclient "github.com/manticoresoftware/manticoresearch-go"
	"github.com/terratensor/library/parser/internal/config"
	"github.com/terratensor/library/parser/internal/library/entry"
	"log"
	"strings"
)

var _ entry.StorageInterface = &Client{}
var apiClient *openapiclient.APIClient

type Client struct {
	apiClient *openapiclient.APIClient
	Index     string
}

func New(ctx context.Context, cfg *config.Manticore) (*Client, error) {
	// Initialize apiClient
	configuration := openapiclient.NewConfiguration()
	configuration.Servers = openapiclient.ServerConfigurations{{URL: serverConfigurationURL(cfg)}}
	apiClient = openapiclient.NewAPIClient(configuration)

	tbl := cfg.Index

	// Check if table exists in cache
	exists, err := tableExists(ctx, apiClient, tbl)
	if err != nil {
		return nil, err
	} else if !exists {
		// Create table if it doesn't exist
		if err = createTable(ctx, apiClient, tbl); err != nil {
			return nil, err
		}
	}

	return &Client{apiClient: apiClient, Index: tbl}, nil
}

func tableExists(ctx context.Context, apiClient *openapiclient.APIClient, tbl string) (bool, error) {

	resp, _, err := apiClient.UtilsAPI.Sql(ctx).Body(fmt.Sprintf("show tables like '%v'", tbl)).Execute()
	if err != nil {
		return false, err
	}
	data := resp[0]["data"].([]interface{})
	log.Println(data)

	return len(data) > 0 && data[0].(map[string]interface{})["Index"] == tbl, nil
}

func createTable(ctx context.Context, apiClient *openapiclient.APIClient, tbl string) error {

	log.Println("Creating table", tbl)
	query := fmt.Sprintf("create table %v(genre text, author text, title text, `text` text, position int, length int) min_infix_len='3' index_exact_words='1' morphology='stem_en, stem_ru' index_sp='1'", tbl)

	sqlRequest := apiClient.UtilsAPI.Sql(ctx).Body(query)
	_, _, err := apiClient.UtilsAPI.SqlExecute(sqlRequest)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Bulk(ctx context.Context, entries []entry.Entry) error {
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
